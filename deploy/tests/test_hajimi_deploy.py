"""Failure-injection tests. Never connect to Docker, SSH, or a real database."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


SCRIPT = Path(__file__).resolve().parents[1] / 'hajimi-deploy.sh'
IMAGE = 'ghcr.io/kule-re/hajimi@sha256:' + 'a' * 64
BASH = os.environ.get('HAJIMI_TEST_BASH') or shutil.which('bash')

FAKE_DOCKER = r'''#!/usr/bin/env bash
set -eu
printf '%s\n' "$*" >> "$FAKE_LOG"
case "$1" in
  inspect)
    case "$4" in
      *working_dir*) printf '%s\n' "${FAKE_WORKING_DIR:-$FAKE_STACK}" ;;
      *config_files*) printf '%s\n' "$FAKE_FILES" ;;
      *'com.docker.compose.project"'*) echo existing-project ;;
      *'.Image'*)
        if [[ -f "$FAKE_STACK/updated" && ${FAKE_FAILURE:-} != mismatch ]]; then
          echo sha256:newimage
        else echo sha256:oldimage; fi ;;
      *) exit 80 ;;
    esac ;;
  login)
    cat > "$FAKE_STACK/token-received"
    printf 'registry-config=%s\n' "$DOCKER_CONFIG" >> "$FAKE_LOG" ;;
  pull) [[ ${FAKE_FAILURE:-} != pull ]] ;;
  image)
    if [[ $2 == inspect ]]; then echo sha256:newimage; fi ;;
  exec)
    [[ ${FAKE_FAILURE:-} != app-data ]] || exit 17
    printf 'fake-app-data-archive' ;;
  compose)
    action=
    for arg in "$@"; do
      case "$arg" in version|config|exec|up|ps) action=$arg; break;; esac
    done
    case "$action" in
      version) echo 'Docker Compose version v2.39.0' ;;
      config) [[ ${FAKE_FAILURE:-} != config ]] ;;
      exec)
        if [[ "$*" == *pg_dump* ]]; then
          [[ ${FAKE_FAILURE:-} != dump ]] || exit 9
          [[ ${FAKE_FAILURE:-} == empty ]] || printf 'PGDMP-test-fixture'
        elif [[ "$*" == *pg_restore* ]]; then
          cat >/dev/null
          [[ ${FAKE_FAILURE:-} != archive ]] || exit 10
        else exit 81; fi ;;
      up)
        touch "$FAKE_STACK/updated"
        [[ ${FAKE_FAILURE:-} != health ]] || exit 12 ;;
      ps) echo 'sub2api healthy' ;;
      *) exit 82 ;;
    esac ;;
  *) exit 83 ;;
esac
'''


class HajimiDeployTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix='hajimi-test-')
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.stack = self.root / 'stack'
        self.stack.mkdir()
        self.bin = self.root / 'bin'
        self.bin.mkdir()
        (self.stack / '.env').write_text('FAKE_CONFIG=keep-me\n', encoding='utf-8')
        self.base = self.stack / 'custom-compose.yml'
        self.base.write_text('services: {}\n', encoding='utf-8')
        self.override = self.stack / 'compose.hajimi.yml'
        self.original_override = 'services:\n  sub2api:\n    image: old-fixture\n'
        self.override.write_text(self.original_override, encoding='utf-8')
        fake = self.bin / 'docker'
        fake.write_text(FAKE_DOCKER, encoding='utf-8', newline='\n')
        fake.chmod(0o755)
        # Git Bash on Windows lacks flock. Production/Ubuntu uses real flock.
        if os.name == 'nt':
            flock = self.bin / 'flock'
            flock.write_text('#!/usr/bin/env bash\nexit 0\n', encoding='utf-8', newline='\n')
            flock.chmod(0o755)
        self.log = self.root / 'calls.txt'
        self.env = {
            **os.environ,
            'FAKE_STACK': self.posix(self.stack),
            'FAKE_FILES': ','.join([self.posix(self.base), self.posix(self.override)]),
            'FAKE_LOG': self.posix(self.log),
            'HAJIMI_TEST_BIN': self.posix(self.bin),
            'HAJIMI_TEST_SCRIPT': self.posix(SCRIPT),
            'HAJIMI_TEST_IMAGE': IMAGE,
        }

    def posix(self, path):
        if os.name != 'nt':
            return str(path)
        return subprocess.check_output(
            [BASH, '-c', 'cygpath -u "$1"', '_', str(path)], text=True, encoding='utf-8'
        ).strip()

    def run_deploy(self, failure='', image=IMAGE, registry=False):
        env = {**self.env, 'FAKE_FAILURE': failure, 'HAJIMI_TEST_IMAGE': image}
        command = 'export PATH="$HAJIMI_TEST_BIN:$PATH"; exec bash "$HAJIMI_TEST_SCRIPT" "$FAKE_STACK" "$HAJIMI_TEST_IMAGE"'
        if registry:
            command += ' fixture-user'
        return subprocess.run(
            [BASH, '-c', command], env=env, text=True, encoding='utf-8',
            input='fixture-token\n' if registry else '',
            # Git Bash process startup is slower under concurrent Windows load.
            stdout=subprocess.PIPE, stderr=subprocess.PIPE,
            timeout=60 if os.name == 'nt' else 20,
        )

    def calls(self):
        return self.log.read_text(encoding='utf-8').splitlines() if self.log.exists() else []

    def assert_not_recreated(self):
        self.assertFalse(any(' up ' in call for call in self.calls()))
        self.assertEqual(self.override.read_text(encoding='utf-8'), self.original_override)
        self.assertEqual((self.stack / '.env').read_text(encoding='utf-8'), 'FAKE_CONFIG=keep-me\n')

    def test_backups_precede_app_only_update_and_original_project_is_preserved(self):
        result = self.run_deploy()
        self.assertEqual(result.returncode, 0, result.stderr + result.stdout)
        calls = self.calls()
        dump = next(i for i, call in enumerate(calls) if 'pg_dump' in call)
        app_data = next(i for i, call in enumerate(calls) if 'exec sub2api tar' in call)
        up = next(i for i, call in enumerate(calls) if ' up ' in call)
        self.assertLess(dump, up)
        self.assertLess(app_data, up)
        self.assertIn('--project-name existing-project', calls[up])
        self.assertIn(self.posix(self.base), calls[up])
        self.assertEqual(calls[up].count(self.posix(self.override)), 1)
        self.assertIn('--no-deps --no-build --pull never --wait --wait-timeout 180 sub2api', calls[up])
        self.assertFalse(any(' down' in call or 'up postgres' in call or 'up redis' in call for call in calls))
        self.assertIn(IMAGE, self.override.read_text(encoding='utf-8'))
        backups = list((self.stack / 'backups').iterdir())
        self.assertEqual(len(backups), 1)
        self.assertEqual((backups[0] / '.env').read_text(encoding='utf-8'), 'FAKE_CONFIG=keep-me\n')
        self.assertTrue((backups[0] / 'database.dump').stat().st_size)
        self.assertTrue((backups[0] / 'app-data.tar.gz').stat().st_size)
        self.assertTrue((backups[0] / 'old-image-tag.txt').exists())

    def test_mutable_or_unexpected_images_are_rejected_before_docker(self):
        for image in ['ghcr.io/kule-re/hajimi:latest', IMAGE.replace('kule-re', 'other')]:
            with self.subTest(image=image):
                self.assertNotEqual(self.run_deploy(image=image).returncode, 0)
                self.assertFalse(self.calls())
                self.assert_not_recreated()

    def test_wrong_working_directory_is_rejected(self):
        other = self.root / 'other'
        other.mkdir()
        self.env['FAKE_WORKING_DIR'] = self.posix(other)
        self.assertNotEqual(self.run_deploy().returncode, 0)
        self.assert_not_recreated()

    def test_missing_existing_env_is_not_bootstrapped(self):
        (self.stack / '.env').unlink()
        self.assertNotEqual(self.run_deploy().returncode, 0)
        self.assertFalse((self.stack / '.env').exists())
        self.assertFalse(any(' up ' in call or call.startswith('pull ') for call in self.calls()))

    def test_missing_original_compose_file_is_rejected(self):
        self.base.unlink()
        self.assertNotEqual(self.run_deploy().returncode, 0)
        self.assert_not_recreated()

    def test_busy_deployment_lock_stops_before_container_mutation(self):
        flock = self.bin / 'flock'
        flock.write_text('#!/usr/bin/env bash\nexit 1\n', encoding='utf-8', newline='\n')
        flock.chmod(0o755)
        result = self.run_deploy()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('another deployment is already running', result.stderr)
        self.assert_not_recreated()

    def test_failures_before_recreation_leave_the_app_and_override_untouched(self):
        for failure in ['pull', 'config', 'dump', 'empty', 'archive', 'app-data']:
            with self.subTest(failure=failure):
                result = self.run_deploy(failure=failure)
                self.assertNotEqual(result.returncode, 0, result.stdout)
                self.assert_not_recreated()
                self.assertFalse(list(self.stack.glob('.hajimi-compose.*.yml')))

    def test_health_failure_keeps_backup_and_does_not_blindly_roll_back(self):
        result = self.run_deploy(failure='health')
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('no automatic image or database rollback', result.stderr)
        self.assertEqual(len([c for c in self.calls() if ' up ' in c]), 1)
        self.assertIn(IMAGE, self.override.read_text(encoding='utf-8'))
        self.assertTrue(list((self.stack / 'backups').glob('*/database.dump')))

    def test_wrong_running_image_is_reported_as_failure(self):
        result = self.run_deploy(failure='mismatch')
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('does not match the requested image', result.stderr)

    def test_registry_token_stays_off_command_line_and_temp_credentials_are_removed(self):
        result = self.run_deploy(registry=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual((self.stack / 'token-received').read_text(encoding='utf-8'), 'fixture-token\n')
        self.assertNotIn('fixture-token', '\n'.join(self.calls()) + result.stdout + result.stderr)
        config_dir = next(c.split('=', 1)[1] for c in self.calls() if c.startswith('registry-config='))
        if os.name == 'nt':
            config_dir = subprocess.check_output([BASH, '-c', 'cygpath -w "$1"', '_', config_dir], text=True, encoding='utf-8').strip()
        self.assertFalse(Path(config_dir).exists())

    def test_repeated_update_does_not_accumulate_generated_overrides(self):
        for _ in range(2):
            result = self.run_deploy()
            self.assertEqual(result.returncode, 0, result.stderr)
        up_calls = [c for c in self.calls() if ' up ' in c]
        self.assertEqual(len(up_calls), 2)
        for call in up_calls:
            self.assertEqual(call.count(self.posix(self.override)), 1)


if __name__ == '__main__':
    unittest.main()
