<template>
<div :class="{ 'is-dark': isDark }" class="wugou-landing bg-background text-foreground relative min-h-svh overflow-x-clip" data-testid="wugou-landing">
 <header class="pointer-events-none fixed inset-x-0 top-0 z-50">
  <div :class="scrolled && !mobileMenuOpen ? 'max-w-[52rem] px-3 pt-3' : 'max-w-7xl px-4 pt-0 md:px-6'" class="pointer-events-auto mx-auto transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]">
   <nav :class="scrolled && !mobileMenuOpen ? 'bg-background/60 ring-border/50 h-12 rounded-2xl pr-1.5 pl-4 shadow-[0_2px_16px_-6px_rgba(0,0,0,0.08),0_0_0_0.5px_rgba(0,0,0,0.02)] ring-[0.5px] backdrop-blur-2xl dark:shadow-[0_2px_16px_-6px_rgba(0,0,0,0.4)]' : 'h-16 px-2'" class="flex items-center justify-between gap-2 transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]">
    <div class="@container/system-brand flex min-w-0 flex-1 items-center gap-1 lg:min-w-36">
     <router-link class="group flex min-w-0 items-center gap-2.5 active" to="/home">
      <div class="flex size-7 shrink-0 items-center justify-center transition-all duration-300 group-hover:scale-105">
       <img :alt="siteName" :src="siteLogo" class="transition-opacity duration-200 opacity-100 size-full rounded-lg object-contain"/>
      </div>
      <span :title="siteName" class="max-w-48 truncate text-sm font-semibold tracking-tight">
       {{ siteName }}
      </span>
     </router-link>
    </div>
    <div class="hidden min-w-0 items-center gap-0.5 lg:flex">
     <router-link class="min-w-0 truncate rounded-lg px-3 py-1.5 text-sm font-medium transition-colors duration-200 text-foreground active" :title="l('主页')" to="/home">
      {{ l('主页') }}
     </router-link>
     <router-link :to="dashboardPath" class="min-w-0 truncate rounded-lg px-3 py-1.5 text-sm font-medium transition-colors duration-200 text-muted-foreground hover:text-foreground" :title="l('控制台')">
      {{ l('控制台') }}
     </router-link>
     <router-link class="min-w-0 truncate rounded-lg px-3 py-1.5 text-sm font-medium transition-colors duration-200 text-muted-foreground hover:text-foreground" :title="l('模型广场')" to="/model-plaza" v-if="showModelPlazaEntry">
      {{ l('模型广场') }}
     </router-link>
     <router-link class="text-muted-foreground hover:text-foreground min-w-0 truncate rounded-lg px-3 py-1.5 text-sm font-medium transition-colors duration-200" :title="l('在线生图')" to="/batch-image">
      {{ l('在线生图') }}
     </router-link>
     <a :href="documentationUrl" class="text-muted-foreground hover:text-foreground min-w-0 truncate rounded-lg px-3 py-1.5 text-sm font-medium transition-colors duration-200" :title="l('文档')">
      {{ l('文档') }}
     </a>
     <div class="bg-border/40 mx-2 h-4 w-px">
     </div>
     <div class="landing-locale" ref="languageMenu">
      <button :aria-expanded="languageMenuOpen" @click="languageMenuOpen = !languageMenuOpen" aria-haspopup="menu" :aria-label="l('更改语言')" class="group/button inline-flex shrink-0 items-center justify-center rounded-lg border border-transparent bg-clip-padding text-sm font-medium whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&amp;_svg]:pointer-events-none [&amp;_svg]:shrink-0 [&amp;_svg:not([class*='size-'])]:size-4 hover:bg-muted hover:text-foreground aria-expanded:bg-muted aria-expanded:text-foreground dark:hover:bg-muted/50 size-8 h-9 w-9" data-slot="dropdown-menu-trigger" tabindex="0" type="button">
       <svg aria-hidden="true" class="lucide lucide-languages size-[1.2rem]" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
        <path d="m5 8 6 6">
        </path>
        <path d="m4 14 6-6 2-3">
        </path>
        <path d="M2 5h12">
        </path>
        <path d="M7 2h1">
        </path>
        <path d="m22 22-5-10-5 10">
        </path>
        <path d="M14 18h6">
        </path>
       </svg>
       <span class="sr-only">
        {{ l('更改语言') }}
       </span>
      </button>
      <div :aria-label="l('语言')" class="landing-language-menu" role="menu" v-if="languageMenuOpen">
       <button :aria-checked="currentLocale === item.code" :key="item.code" @click="changeLanguage(item.code)" role="menuitemradio" type="button" v-for="item in availableLocales">
        <span>
         {{ item.name }}
        </span>
        <span aria-hidden="true" v-if="currentLocale === item.code">
         ✓
        </span>
       </button>
      </div>
     </div>
     <button :aria-pressed="isDark" :title="isDark ? l('切换浅色主题') : l('切换深色主题')" @click="emit('toggle-theme')" :aria-label="l('切换主题')" class="group/button inline-flex shrink-0 items-center justify-center rounded-lg border border-transparent bg-clip-padding text-sm font-medium whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&amp;_svg]:pointer-events-none [&amp;_svg]:shrink-0 [&amp;_svg:not([class*='size-'])]:size-4 hover:bg-muted hover:text-foreground aria-expanded:bg-muted aria-expanded:text-foreground dark:hover:bg-muted/50 size-8 h-9 w-9" data-slot="button" tabindex="0" type="button">
      <svg v-if="isDark" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-moon size-5" aria-hidden="true"><path d="M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401"></path></svg>
<svg v-else aria-hidden="true" class="lucide lucide-sun size-5" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
       <circle cx="12" cy="12" r="4">
       </circle>
       <path d="M12 2v2">
       </path>
       <path d="M12 20v2">
       </path>
       <path d="m4.93 4.93 1.41 1.41">
       </path>
       <path d="m17.66 17.66 1.41 1.41">
       </path>
       <path d="M2 12h2">
       </path>
       <path d="M20 12h2">
       </path>
       <path d="m6.34 17.66-1.41 1.41">
       </path>
       <path d="m19.07 4.93-1.41 1.41">
       </path>
      </svg>
     </button>
     <AnnouncementBell v-if="isAuthenticated">
     </AnnouncementBell>
     <router-link :aria-label="l('登录后查看公告')" class="group/button inline-flex shrink-0 items-center justify-center rounded-lg border border-transparent bg-clip-padding text-sm font-medium whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&amp;_svg]:pointer-events-none [&amp;_svg]:shrink-0 [&amp;_svg:not([class*='size-'])]:size-4 hover:bg-muted hover:text-foreground aria-expanded:bg-muted aria-expanded:text-foreground dark:hover:bg-muted/50 relative size-9" data-slot="popover-trigger" tabindex="0" :title="l('登录后查看公告')" to="/login" v-else>
      <svg aria-hidden="true" class="lucide lucide-bell size-[1.2rem]" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
       <path d="M10.268 21a2 2 0 0 0 3.464 0">
       </path>
       <path d="M3.262 15.326A1 1 0 0 0 4 17h16a1 1 0 0 0 .74-1.673C19.41 13.956 18 12.499 18 8A6 6 0 0 0 6 8c0 4.499-1.411 5.956-2.738 7.326">
       </path>
      </svg>
     </router-link>
     <div class="bg-border/40 mx-1 h-4 w-px">
     </div>
     <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="group/button inline-flex shrink-0 items-center justify-center border border-transparent bg-clip-padding whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&amp;_svg]:pointer-events-none [&amp;_svg]:shrink-0 bg-primary text-primary-foreground [a]:hover:bg-primary/80 gap-1 in-data-[slot=button-group]:rounded-lg has-data-[icon=inline-end]:pr-1.5 has-data-[icon=inline-start]:pl-1.5 [&amp;_svg:not([class*='size-'])]:size-3.5 h-8 rounded-lg px-3.5 text-xs font-medium" data-slot="button" tabindex="0">
      {{ isAuthenticated ? l('控制台') : l('登录') }}
     </router-link>
    </div>
    <div class="flex shrink-0 items-center gap-2 lg:hidden">
     <button :aria-pressed="isDark" :title="isDark ? l('切换浅色主题') : l('切换深色主题')" @click="emit('toggle-theme')" :aria-label="l('切换主题')" class="group/button inline-flex shrink-0 items-center justify-center rounded-lg border border-transparent bg-clip-padding text-sm font-medium whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&amp;_svg]:pointer-events-none [&amp;_svg]:shrink-0 [&amp;_svg:not([class*='size-'])]:size-4 hover:bg-muted hover:text-foreground aria-expanded:bg-muted aria-expanded:text-foreground dark:hover:bg-muted/50 size-8 h-9 w-9" data-slot="button" tabindex="0" type="button">
      <svg v-if="isDark" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-moon size-5" aria-hidden="true"><path d="M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401"></path></svg>
<svg v-else aria-hidden="true" class="lucide lucide-sun size-5" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
       <circle cx="12" cy="12" r="4">
       </circle>
       <path d="M12 2v2">
       </path>
       <path d="M12 20v2">
       </path>
       <path d="m4.93 4.93 1.41 1.41">
       </path>
       <path d="m17.66 17.66 1.41 1.41">
       </path>
       <path d="M2 12h2">
       </path>
       <path d="M20 12h2">
       </path>
       <path d="m6.34 17.66-1.41 1.41">
       </path>
       <path d="m19.07 4.93-1.41 1.41">
       </path>
      </svg>
     </button>
     <button :aria-expanded="mobileMenuOpen" @click="mobileMenuOpen = !mobileMenuOpen" aria-controls="landing-mobile-menu" :aria-label="l('切换导航菜单')" class="group/button inline-flex shrink-0 items-center justify-center rounded-lg border border-transparent bg-clip-padding text-sm font-medium whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&amp;_svg]:pointer-events-none [&amp;_svg]:shrink-0 [&amp;_svg:not([class*='size-'])]:size-4 hover:bg-muted hover:text-foreground aria-expanded:bg-muted aria-expanded:text-foreground dark:hover:bg-muted/50 size-9" data-slot="button" ref="mobileMenuButton" tabindex="0" type="button">
      <div class="relative size-4">
       <span :class="mobileMenuOpen ? 'top-[7px] rotate-45' : 'top-[3px]'" class="absolute inset-x-0 block h-[1.5px] origin-center rounded-full bg-current transition-all duration-300">
       </span>
       <span :class="mobileMenuOpen ? 'opacity-0' : 'opacity-100'" class="absolute inset-x-0 top-[7px] block h-[1.5px] rounded-full bg-current transition-all duration-300">
       </span>
       <span :class="mobileMenuOpen ? 'top-[7px] -rotate-45' : 'top-[11px]'" class="absolute inset-x-0 block h-[1.5px] origin-center rounded-full bg-current transition-all duration-300">
       </span>
      </div>
     </button>
    </div>
   </nav>
  </div>
 </header>
 <div :aria-label="l('网站导航')" aria-modal="true" class="bg-background/98 fixed inset-0 z-40 backdrop-blur-2xl transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] lg:pointer-events-none lg:hidden" id="landing-mobile-menu" ref="mobileMenu" role="dialog" v-if="mobileMenuOpen">
  <div class="flex h-full flex-col justify-between px-8 pt-20 pb-10">
   <nav class="flex flex-col gap-1">
    <router-link @click="mobileMenuOpen = false" class="flex items-center gap-3 py-3 text-base font-medium tracking-tight transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] text-foreground active" to="/home">
     {{ l('主页') }}
    </router-link>
    <router-link :to="dashboardPath" @click="mobileMenuOpen = false" class="flex items-center gap-3 py-3 text-base font-medium tracking-tight transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] text-muted-foreground">
     {{ l('控制台') }}
    </router-link>
    <router-link @click="mobileMenuOpen = false" class="flex items-center gap-3 py-3 text-base font-medium tracking-tight transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] text-muted-foreground" to="/model-plaza" v-if="showModelPlazaEntry">
     {{ l('模型广场') }}
    </router-link>
    <router-link @click="mobileMenuOpen = false" class="flex items-center gap-3 py-3 text-base font-medium tracking-tight transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] text-muted-foreground" to="/batch-image">
     {{ l('在线生图') }}
    </router-link>
    <a :href="documentationUrl" @click="mobileMenuOpen = false" class="flex items-center gap-3 py-3 text-base font-medium tracking-tight transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] text-muted-foreground">
     {{ l('文档') }}
    </a>
   </nav>
   <div class="flex flex-col gap-3 transition-all duration-500">
    <router-link :to="isAuthenticated ? dashboardPath : '/login'" @click="mobileMenuOpen = false" class="bg-foreground text-background inline-flex h-10 items-center justify-center rounded-lg text-sm font-medium transition-opacity hover:opacity-90 active:opacity-80">
     {{ isAuthenticated ? l('控制台') : l('登录') }}
    </router-link>
   </div>
  </div>
 </div>
 <section class="relative z-10 overflow-hidden px-6 pt-24 pb-16 md:pt-32 md:pb-24 lg:pt-36 lg:pb-28">
  <div aria-hidden="true" class="pointer-events-none absolute inset-0 -z-10 opacity-25 dark:opacity-[0.12]" style="background: radial-gradient(60% 50% at 20% 20%, oklch(0.72 0.18 250 / 0.8) 0%, transparent 70%), radial-gradient(50% 40% at 80% 15%, oklch(0.65 0.15 200 / 0.6) 0%, transparent 70%), radial-gradient(40% 35% at 40% 80%, oklch(0.7 0.12 280 / 0.4) 0%, transparent 70%);">
  </div>
  <div aria-hidden="true" class="absolute inset-0 -z-10 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_30%,black_20%,transparent_100%)] bg-[size:4rem_4rem] opacity-[0.08]">
  </div>
  <div class="mx-auto grid max-w-6xl grid-cols-1 items-start gap-12 lg:grid-cols-12 lg:gap-8">
   <div class="flex flex-col items-start text-left lg:col-span-6">
    <h1 class="landing-animate-fade-up text-[clamp(2.25rem,4.5vw,3.25rem)] leading-[1.15] font-bold tracking-tight" style="animation-delay: 60ms;">
     {{ l('统一 API 网关，服务于') }}
     <br/>
     <span class="bg-gradient-to-r from-blue-400 via-violet-400 to-purple-500 bg-clip-text text-transparent">
      {{ l('海量 AI 模型') }}
     </span>
    </h1>
    <p class="landing-animate-fade-up text-muted-foreground/80 mt-5 max-w-xl text-base leading-relaxed md:text-[15px]" style="animation-delay: 120ms;">
     {{ l('通过统一、标准的接口协议接入海量模型。承载 AI 应用，高效管理数字资产，连接未来。') }}
    </p>
    <div class="landing-animate-fade-up mt-8 flex flex-wrap items-center gap-3" style="animation-delay: 180ms;">
     <router-link :to="startPath" class="group/button inline-flex shrink-0 items-center justify-center border border-transparent bg-clip-padding whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&amp;_svg]:pointer-events-none [&amp;_svg]:shrink-0 [&amp;_svg:not([class*='size-'])]:size-4 bg-primary text-primary-foreground [a]:hover:bg-primary/80 gap-1.5 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2 group h-11 rounded-lg px-5 text-sm font-medium" data-slot="button" tabindex="0">
      {{ l('开始使用') }}
      <svg aria-hidden="true" class="lucide lucide-arrow-right ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
       <path d="M5 12h14">
       </path>
       <path d="m12 5 7 7-7 7">
       </path>
      </svg>
     </router-link>
     <router-link :to="pricingPath" class="group/button inline-flex shrink-0 items-center justify-center border bg-clip-padding whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&amp;_svg]:pointer-events-none [&amp;_svg]:shrink-0 [&amp;_svg:not([class*='size-'])]:size-4 bg-background hover:text-foreground aria-expanded:bg-muted aria-expanded:text-foreground dark:border-input dark:bg-input/30 dark:hover:bg-input/50 gap-1.5 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2 border-border/50 hover:border-border hover:bg-muted/50 h-11 rounded-lg px-5 text-sm font-medium" data-slot="button" tabindex="0">
      {{ l('查看定价') }}
     </router-link>
     <a :href="documentationUrl" class="group/button shrink-0 justify-center border bg-clip-padding whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&amp;_svg]:pointer-events-none [&amp;_svg]:shrink-0 [&amp;_svg:not([class*='size-'])]:size-4 bg-background hover:text-foreground aria-expanded:bg-muted aria-expanded:text-foreground dark:border-input dark:bg-input/30 dark:hover:bg-input/50 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2 group border-border/50 hover:border-border hover:bg-muted/50 inline-flex h-11 items-center gap-1.5 rounded-lg px-5 text-sm font-medium" data-slot="button" role="button" tabindex="0">
      <svg aria-hidden="true" class="lucide lucide-book-open text-muted-foreground/80 group-hover:text-foreground size-4 transition-colors duration-200" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
       <path d="M12 7v14">
       </path>
       <path d="M3 18a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h5a4 4 0 0 1 4 4 4 4 0 0 1 4-4h5a1 1 0 0 1 1 1v13a1 1 0 0 1-1 1h-6a3 3 0 0 0-3 3 3 3 0 0 0-3-3z">
       </path>
      </svg>
      <span>
       {{ l('文档') }}
      </span>
     </a>
    </div>
    <div class="landing-animate-fade-up mt-10 w-full max-w-xl" style="animation-delay: 240ms;">
     <div class="mb-4 flex flex-col gap-1">
      <span class="text-muted-foreground/50 text-[10px] font-bold tracking-[0.15em] uppercase">
       {{ l('常用应用支持') }}
      </span>
      <p class="text-muted-foreground/60 text-xs leading-relaxed">
       {{ l('通过{siteName} 接入常用 AI 应用与开发工具') }}
      </p>
     </div>
     <div class="flex flex-wrap items-center gap-3">
      <a class="group border-border/40 bg-muted/15 text-foreground/80 hover:border-border hover:bg-muted/30 hover:text-foreground flex items-center gap-3 rounded-full border px-5 py-2.5 text-sm font-medium shadow-[0_1px_2.5px_rgba(0,0,0,0.01)] backdrop-blur-xs transition-all duration-300 hover:scale-[1.02]" href="https://www.workbuddy.cn/" rel="noopener noreferrer" target="_blank" title="WorkBuddy">
       <span aria-hidden="true" class="size-6 shrink-0 rounded-md bg-blue-500/10 text-[10px] font-bold text-blue-600 dark:bg-blue-400/10 dark:text-blue-400" style="display: grid; place-items: center;">
        WB
       </span>
       <span>
        WorkBuddy
       </span>
      </a>
      <a class="group border-border/40 bg-muted/15 text-foreground/80 hover:border-border hover:bg-muted/30 hover:text-foreground flex items-center gap-3 rounded-full border px-5 py-2.5 text-sm font-medium shadow-[0_1px_2.5px_rgba(0,0,0,0.01)] backdrop-blur-xs transition-all duration-300 hover:scale-[1.02]" href="https://ccswitch.io" rel="noopener noreferrer" target="_blank">
       <img alt="CC Switch" class="size-6 shrink-0 rounded-md object-contain" src="/landing/cc-switch.png"/>
       <span class="size-6 shrink-0 items-center justify-center rounded-md bg-blue-500/10 text-[10px] font-bold text-blue-600 dark:bg-blue-400/10 dark:text-blue-400" style="display: none;">
        CC
       </span>
       <span>
        CC Switch
       </span>
      </a>
      <router-link class="group border-border/40 bg-muted/15 text-foreground/55 hover:border-border hover:bg-muted/30 hover:text-foreground flex items-center gap-2.5 rounded-full border px-5 py-2.5 text-sm font-medium shadow-[0_1px_2.5px_rgba(0,0,0,0.01)] backdrop-blur-xs transition-all duration-300 hover:scale-[1.02]" :title="l('WorkBuddy 配置指南')" to="/workbuddy">
       <svg class="text-muted-foreground/60 group-hover:text-foreground size-6 shrink-0 transition-colors" fill="none" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
        <circle cx="6" cy="12" fill="currentColor" r="2">
        </circle>
        <circle cx="12" cy="12" fill="currentColor" r="2">
        </circle>
        <circle cx="18" cy="12" fill="currentColor" r="2">
        </circle>
       </svg>
       <span>
        {{ l('更多') }}
       </span>
      </router-link>
     </div>
    </div>
   </div>
   <div class="landing-animate-fade-up flex w-full justify-center lg:col-span-6" style="animation-delay: 320ms;">
    <div class="mx-auto w-full max-w-2xl mt-8 lg:mt-0">
     <div :aria-label="l('接口调用示例（静态演示）')" class="overflow-hidden rounded-2xl border backdrop-blur-sm border-border/60 bg-white/95 shadow-[0_20px_50px_-25px_rgba(15,23,42,0.18)] dark:border-white/[0.06] dark:bg-[#0b0f17]/95 dark:shadow-[0_20px_60px_-25px_rgba(0,0,0,0.7)]" data-testid="protocol-example">
      <div class="flex items-center gap-1 border-b px-2 sm:gap-1.5 sm:px-3 border-border/50 dark:border-white/[0.05]">
       <div @keydown="onProtocolKeydown" :aria-label="l('API 协议')" class="landing-protocol-tabs" role="tablist">
        <button :aria-selected="activeProtocol === index" :class="activeProtocol === index ? protocol.activeClass : 'text-foreground/40 hover:text-foreground/70 border-transparent'" :id="'protocol-tab-' + protocol.id" :key="protocol.id" :tabindex="activeProtocol === index ? 0 : -1" @click="selectProtocol(index)" aria-controls="protocol-panel" class="relative -mb-px flex items-center gap-1.5 border-b-2 px-2.5 py-2.5 text-[11px] font-medium tracking-wide transition-colors sm:px-3 sm:text-xs" role="tab" type="button" v-for="(protocol, index) in landingProtocols">
         {{ protocol.label }}
        </button>
       </div>
       <div class="ml-auto flex items-center gap-2 pr-2 sm:pr-3">
        <span class="inline-block size-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.45)]">
        </span>
        <span class="text-foreground/40 font-mono text-[10px] tracking-wider uppercase">
         200 ok
        </span>
       </div>
      </div>
      <div :aria-labelledby="'protocol-tab-' + landingProtocols[activeProtocol].id" id="protocol-panel" role="tabpanel" v-html="landingProtocols[activeProtocol].html">
      </div>
     </div>
    </div>
   </div>
  </div>
 </section>
 <div class="border-border/40 bg-muted/10 relative z-10 border-y">
  <div class="mx-auto max-w-6xl px-6 py-10 md:py-12">
   <div class="grid grid-cols-2 gap-8 md:grid-cols-4 md:gap-12">
    <div class="flex flex-col items-center text-center">
     <span class="text-2xl font-bold tracking-tight md:text-3xl">
      <span class="tabular-nums">
       {{ l('多厂商') }}
      </span>
     </span>
     <span class="text-muted-foreground mt-1.5 text-xs">
      {{ l('模型统一接入') }}
     </span>
    </div>
    <div class="flex flex-col items-center text-center">
     <span class="text-2xl font-bold tracking-tight md:text-3xl">
      <span class="tabular-nums">
       {{ l('按量') }}
      </span>
     </span>
     <span class="text-muted-foreground mt-1.5 text-xs">
      {{ l('用量与费用记录') }}
     </span>
    </div>
    <div class="flex flex-col items-center text-center">
     <span class="text-2xl font-bold tracking-tight md:text-3xl">
      <span class="tabular-nums">
       {{ l('多协议') }}
      </span>
     </span>
     <span class="text-muted-foreground mt-1.5 text-xs">
      {{ l('兼容 API 调用') }}
     </span>
    </div>
    <div class="flex flex-col items-center text-center">
     <span class="text-2xl font-bold tracking-tight md:text-3xl">
      <span class="tabular-nums">
       {{ l('分组') }}
      </span>
     </span>
     <span class="text-muted-foreground mt-1.5 text-xs">
      {{ l('权限与调度管理') }}
     </span>
    </div>
   </div>
  </div>
 </div>
 <section class="relative z-10 px-6 py-24 md:py-32">
  <div class="mx-auto max-w-6xl">
   <div class="mb-16 max-w-lg">
    <p class="text-muted-foreground mb-3 text-xs font-medium tracking-widest uppercase">
     {{ l('核心功能') }}
    </p>
    <h2 class="text-2xl leading-tight font-bold tracking-tight md:text-3xl">
     {{ l('为开发者打造，') }}
     <br/>
     {{ l('为规模而设计') }}
    </h2>
   </div>
   <div class="border-border/40 bg-border/40 grid gap-px overflow-hidden rounded-xl border md:grid-cols-3">
    <div class="bg-background group hover:bg-muted/20 p-7 transition-colors duration-300 md:p-8 md:col-span-2">
     <div class="mb-3 flex items-center gap-3">
      <span class="border-border/40 bg-muted text-muted-foreground flex size-7 items-center justify-center rounded-md border text-[10px] font-semibold tabular-nums">
       01
      </span>
      <h3 class="text-sm font-semibold">
       {{ l('极速') }}
      </h3>
     </div>
     <p class="text-muted-foreground text-sm leading-relaxed">
      {{ l('统一接入多家模型，减少应用适配工作') }}
     </p>
     <div class="mt-4 grid grid-cols-3 gap-2">
      <div class="border-border/30 bg-muted/20 text-muted-foreground flex items-center justify-center rounded-lg border px-3 py-2 text-xs transition-colors duration-300 hover:border-blue-500/30 hover:bg-blue-500/5">
       OpenAI
      </div>
      <div class="border-border/30 bg-muted/20 text-muted-foreground flex items-center justify-center rounded-lg border px-3 py-2 text-xs transition-colors duration-300 hover:border-blue-500/30 hover:bg-blue-500/5">
       Claude
      </div>
      <div class="border-border/30 bg-muted/20 text-muted-foreground flex items-center justify-center rounded-lg border px-3 py-2 text-xs transition-colors duration-300 hover:border-blue-500/30 hover:bg-blue-500/5">
       Gemini
      </div>
      <div class="border-border/30 bg-muted/20 text-muted-foreground flex items-center justify-center rounded-lg border px-3 py-2 text-xs transition-colors duration-300 hover:border-blue-500/30 hover:bg-blue-500/5">
       DeepSeek
      </div>
      <div class="border-border/30 bg-muted/20 text-muted-foreground flex items-center justify-center rounded-lg border px-3 py-2 text-xs transition-colors duration-300 hover:border-blue-500/30 hover:bg-blue-500/5">
       Qwen
      </div>
      <div class="border-border/30 bg-muted/20 text-muted-foreground flex items-center justify-center rounded-lg border px-3 py-2 text-xs transition-colors duration-300 hover:border-blue-500/30 hover:bg-blue-500/5">
       Llama
      </div>
     </div>
    </div>
    <div class="bg-background group hover:bg-muted/20 p-7 transition-colors duration-300 md:p-8 md:col-span-1" style="animation-delay: 100ms;">
     <div class="mb-3 flex items-center gap-3">
      <span class="border-border/40 bg-muted text-muted-foreground flex size-7 items-center justify-center rounded-md border text-[10px] font-semibold tabular-nums">
       02
      </span>
      <h3 class="text-sm font-semibold">
       {{ l('安全可靠') }}
      </h3>
     </div>
     <p class="text-muted-foreground text-sm leading-relaxed">
      {{ l('API 密钥与权限分组，管理用户访问') }}
     </p>
     <div class="mt-4 flex items-center justify-center">
      <div class="relative">
       <div class="flex size-16 items-center justify-center rounded-2xl border border-emerald-500/20 bg-emerald-500/5">
        <svg aria-hidden="true" class="lucide lucide-shield size-7 text-emerald-500/70" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
         <path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z">
         </path>
        </svg>
       </div>
       <div class="absolute -top-1 -right-1 flex size-4 items-center justify-center rounded-full bg-emerald-500">
        <svg class="size-2.5 text-white" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24">
         <path d="m4.5 12.75 6 6 9-13.5" stroke-linecap="round" stroke-linejoin="round">
         </path>
        </svg>
       </div>
      </div>
     </div>
    </div>
    <div class="bg-background group hover:bg-muted/20 p-7 transition-colors duration-300 md:p-8 md:col-span-1" style="animation-delay: 200ms;">
     <div class="mb-3 flex items-center gap-3">
      <span class="border-border/40 bg-muted text-muted-foreground flex size-7 items-center justify-center rounded-md border text-[10px] font-semibold tabular-nums">
       03
      </span>
      <h3 class="text-sm font-semibold">
       {{ l('稳定调度') }}
      </h3>
     </div>
     <p class="text-muted-foreground text-sm leading-relaxed">
      {{ l('分组与限流，统一管理上游请求') }}
     </p>
     <div class="mt-4 space-y-2">
      <div class="flex items-center gap-2">
       <div class="flex size-6 items-center justify-center rounded-full text-[10px] font-bold border-border/40 bg-muted text-muted-foreground border">
        1
       </div>
       <div class="bg-border/40 h-px flex-1">
       </div>
       <span class="text-muted-foreground text-xs">
        {{ l('负载均衡') }}
       </span>
      </div>
      <div class="flex items-center gap-2">
       <div class="flex size-6 items-center justify-center rounded-full text-[10px] font-bold border border-blue-500/30 bg-blue-500/20 text-blue-500">
        2
       </div>
       <div class="bg-border/40 h-px flex-1">
       </div>
       <span class="text-muted-foreground text-xs">
        {{ l('速率限制') }}
       </span>
      </div>
      <div class="flex items-center gap-2">
       <div class="flex size-6 items-center justify-center rounded-full text-[10px] font-bold border-border/40 bg-muted text-muted-foreground border">
        3
       </div>
       <div class="bg-border/40 h-px flex-1">
       </div>
       <span class="text-muted-foreground text-xs">
        {{ l('成本跟踪') }}
       </span>
      </div>
     </div>
    </div>
    <div class="bg-background group hover:bg-muted/20 p-7 transition-colors duration-300 md:p-8 md:col-span-2" style="animation-delay: 300ms;">
     <div class="mb-3 flex items-center gap-3">
      <span class="border-border/40 bg-muted text-muted-foreground flex size-7 items-center justify-center rounded-md border text-[10px] font-semibold tabular-nums">
       04
      </span>
      <h3 class="text-sm font-semibold">
       {{ l('开发者友好') }}
      </h3>
     </div>
     <p class="text-muted-foreground text-sm leading-relaxed">
      {{ l('兼容常见 AI 应用工作流的 API 路由') }}
     </p>
     <div class="mt-4 flex items-center gap-3">
      <div class="flex -space-x-2">
       <div class="border-background from-muted to-muted/60 text-muted-foreground flex size-8 items-center justify-center rounded-full border-2 bg-gradient-to-br text-[9px] font-bold">
        API
       </div>
       <div class="border-background from-muted to-muted/60 text-muted-foreground flex size-8 items-center justify-center rounded-full border-2 bg-gradient-to-br text-[9px] font-bold">
        SDK
       </div>
       <div class="border-background from-muted to-muted/60 text-muted-foreground flex size-8 items-center justify-center rounded-full border-2 bg-gradient-to-br text-[9px] font-bold">
        CLI
       </div>
       <div class="border-background from-muted to-muted/60 text-muted-foreground flex size-8 items-center justify-center rounded-full border-2 bg-gradient-to-br text-[9px] font-bold">
        Docs
       </div>
      </div>
      <div class="text-muted-foreground flex items-center gap-1.5 text-xs">
       <svg aria-hidden="true" class="lucide lucide-code size-3.5 text-blue-500" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
        <path d="m16 18 6-6-6-6">
        </path>
        <path d="m8 6-6 6 6 6">
        </path>
       </svg>
       {{ l('兼容多协议') }}
      </div>
     </div>
    </div>
   </div>
   <div class="mt-12 grid grid-cols-2 gap-8 md:grid-cols-4 md:gap-12">
    <div class="flex flex-col items-center text-center">
     <div class="text-muted-foreground border-border/50 bg-muted/30 group-hover:text-foreground mb-3 flex size-12 items-center justify-center rounded-xl border transition-colors">
      <svg aria-hidden="true" class="lucide lucide-gauge size-5" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
       <path d="m12 14 4-4">
       </path>
       <path d="M3.34 19a10 10 0 1 1 17.32 0">
       </path>
      </svg>
     </div>
     <h3 class="mb-1.5 text-sm font-semibold">
      {{ l('高性能') }}
     </h3>
     <p class="text-muted-foreground max-w-[200px] text-xs leading-relaxed">
      {{ l('支持并发控制和账号调度') }}
     </p>
    </div>
    <div class="flex flex-col items-center text-center" style="animation-delay: 100ms;">
     <div class="text-muted-foreground border-border/50 bg-muted/30 group-hover:text-foreground mb-3 flex size-12 items-center justify-center rounded-xl border transition-colors">
      <svg aria-hidden="true" class="lucide lucide-dollar-sign size-5" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
       <line x1="12" x2="12" y1="2" y2="22">
       </line>
       <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6">
       </path>
      </svg>
     </div>
     <h3 class="mb-1.5 text-sm font-semibold">
      {{ l('透明计费') }}
     </h3>
     <p class="text-muted-foreground max-w-[200px] text-xs leading-relaxed">
      {{ l('按量付费，实时监控使用情况') }}
     </p>
    </div>
    <div class="flex flex-col items-center text-center" style="animation-delay: 200ms;">
     <div class="text-muted-foreground border-border/50 bg-muted/30 group-hover:text-foreground mb-3 flex size-12 items-center justify-center rounded-xl border transition-colors">
      <svg aria-hidden="true" class="lucide lucide-users size-5" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
       <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2">
       </path>
       <path d="M16 3.128a4 4 0 0 1 0 7.744">
       </path>
       <path d="M22 21v-2a4 4 0 0 0-3-3.87">
       </path>
       <circle cx="9" cy="7" r="4">
       </circle>
      </svg>
     </div>
     <h3 class="mb-1.5 text-sm font-semibold">
      {{ l('团队协作') }}
     </h3>
     <p class="text-muted-foreground max-w-[200px] text-xs leading-relaxed">
      {{ l('用户与分组管理，灵活分配权限') }}
     </p>
    </div>
    <div class="flex flex-col items-center text-center" style="animation-delay: 300ms;">
     <div class="text-muted-foreground border-border/50 bg-muted/30 group-hover:text-foreground mb-3 flex size-12 items-center justify-center rounded-xl border transition-colors">
      <svg aria-hidden="true" class="lucide lucide-heart-handshake size-5" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
       <path d="M19.414 14.414C21 12.828 22 11.5 22 9.5a5.5 5.5 0 0 0-9.591-3.676.6.6 0 0 1-.818.001A5.5 5.5 0 0 0 2 9.5c0 2.3 1.5 4 3 5.5l5.535 5.362a2 2 0 0 0 2.879.052 2.12 2.12 0 0 0-.004-3 2.124 2.124 0 1 0 3-3 2.124 2.124 0 0 0 3.004 0 2 2 0 0 0 0-2.828l-1.881-1.882a2.41 2.41 0 0 0-3.409 0l-1.71 1.71a2 2 0 0 1-2.828 0 2 2 0 0 1 0-2.828l2.823-2.762">
       </path>
      </svg>
     </div>
     <h3 class="mb-1.5 text-sm font-semibold">
      {{ l('灵活扩展') }}
     </h3>
     <p class="text-muted-foreground max-w-[200px] text-xs leading-relaxed">
      {{ l('{siteName}支持多种 API 协议与应用接入') }}
     </p>
    </div>
   </div>
  </div>
 </section>
 <section class="border-border/40 relative z-10 border-t px-6 py-24 md:py-32">
  <div class="mx-auto max-w-6xl">
   <div class="mb-16 text-center md:mb-20">
    <p class="text-muted-foreground mb-3 text-xs font-medium tracking-widest uppercase">
     {{ l('工作流程') }}
    </p>
    <h2 class="text-2xl font-bold tracking-tight md:text-3xl">
     {{ l('三步快速上手') }}
    </h2>
   </div>
   <div class="grid gap-8 md:grid-cols-3 md:gap-12">
    <div class="relative flex flex-col items-center text-center">
     <div class="relative mb-6">
      <div class="text-muted-foreground border-border/50 bg-muted/30 flex size-16 items-center justify-center rounded-2xl border transition-colors">
       <svg aria-hidden="true" class="lucide lucide-settings size-6" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
        <path d="M9.671 4.136a2.34 2.34 0 0 1 4.659 0 2.34 2.34 0 0 0 3.319 1.915 2.34 2.34 0 0 1 2.33 4.033 2.34 2.34 0 0 0 0 3.831 2.34 2.34 0 0 1-2.33 4.033 2.34 2.34 0 0 0-3.319 1.915 2.34 2.34 0 0 1-4.659 0 2.34 2.34 0 0 0-3.32-1.915 2.34 2.34 0 0 1-2.33-4.033 2.34 2.34 0 0 0 0-3.831A2.34 2.34 0 0 1 6.35 6.051a2.34 2.34 0 0 0 3.319-1.915">
        </path>
        <circle cx="12" cy="12" r="3">
        </circle>
       </svg>
      </div>
      <div class="bg-foreground text-background absolute -top-2 -right-2 flex size-6 items-center justify-center rounded-full text-xs font-bold">
       1
      </div>
     </div>
     <h3 class="mb-2 text-base font-semibold">
      {{ l('配置') }}
     </h3>
     <p class="text-muted-foreground max-w-[240px] text-sm leading-relaxed">
      {{ l('注册账号，创建你的 API 密钥并选择可用模型') }}
     </p>
    </div>
    <div class="relative flex flex-col items-center text-center" style="animation-delay: 150ms;">
     <div class="relative mb-6">
      <div class="text-muted-foreground border-border/50 bg-muted/30 flex size-16 items-center justify-center rounded-2xl border transition-colors">
       <svg aria-hidden="true" class="lucide lucide-zap size-6" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
        <path d="M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z">
        </path>
       </svg>
      </div>
      <div class="bg-foreground text-background absolute -top-2 -right-2 flex size-6 items-center justify-center rounded-full text-xs font-bold">
       2
      </div>
     </div>
     <h3 class="mb-2 text-base font-semibold">
      {{ l('连接') }}
     </h3>
     <p class="text-muted-foreground max-w-[240px] text-sm leading-relaxed">
      {{ l('通过 OpenAI、Claude、Gemini 以及其他兼容 API 路由接入') }}
     </p>
    </div>
    <div class="relative flex flex-col items-center text-center" style="animation-delay: 300ms;">
     <div class="relative mb-6">
      <div class="text-muted-foreground border-border/50 bg-muted/30 flex size-16 items-center justify-center rounded-2xl border transition-colors">
       <svg aria-hidden="true" class="lucide lucide-chart-column size-6" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
        <path d="M3 3v16a2 2 0 0 0 2 2h16">
        </path>
        <path d="M18 17V9">
        </path>
        <path d="M13 17V5">
        </path>
        <path d="M8 17v-3">
        </path>
       </svg>
      </div>
      <div class="bg-foreground text-background absolute -top-2 -right-2 flex size-6 items-center justify-center rounded-full text-xs font-bold">
       3
      </div>
     </div>
     <h3 class="mb-2 text-base font-semibold">
      {{ l('监控') }}
     </h3>
     <p class="text-muted-foreground max-w-[240px] text-sm leading-relaxed">
      {{ l('通过实时分析跟踪用量、成本和性能') }}
     </p>
    </div>
   </div>
  </div>
 </section>
 <section class="relative z-10 overflow-hidden px-6 py-24 md:py-32">
  <div aria-hidden="true" class="absolute inset-0 -z-10 opacity-20 dark:opacity-[0.08]" style="background: radial-gradient(50% 50% at 30% 50%, oklch(0.7 0.15 250 / 0.7) 0%, transparent 70%), radial-gradient(40% 40% at 70% 40%, oklch(0.65 0.12 200 / 0.5) 0%, transparent 70%);">
  </div>
  <div class="mx-auto max-w-2xl text-center">
   <h2 class="text-2xl leading-tight font-bold tracking-tight md:text-4xl">
    {{ l('准备好简化') }}
    <br/>
    <span class="bg-gradient-to-r from-blue-400 via-violet-400 to-purple-500 bg-clip-text text-transparent">
     {{ l('你的 AI 集成了吗？') }}
    </span>
   </h2>
   <p class="text-muted-foreground/80 mx-auto mt-5 max-w-md text-sm leading-relaxed md:text-base">
    {{ l('在{siteName} 创建 API 密钥，将本站模型接入你的应用与开发工具。') }}
   </p>
   <div class="mt-8 flex items-center justify-center gap-3">
    <router-link :to="startPath" class="group/button inline-flex shrink-0 items-center justify-center border border-transparent bg-clip-padding text-sm font-medium whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&amp;_svg]:pointer-events-none [&amp;_svg]:shrink-0 [&amp;_svg:not([class*='size-'])]:size-4 bg-primary text-primary-foreground [a]:hover:bg-primary/80 h-8 gap-1.5 px-2.5 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2 group rounded-lg" data-slot="button" tabindex="0">
     {{ l('开始使用') }}
     <svg aria-hidden="true" class="lucide lucide-arrow-right ml-1 size-3.5 transition-transform duration-200 group-hover:translate-x-0.5" fill="none" height="24" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg">
      <path d="M5 12h14">
      </path>
      <path d="m12 5 7 7-7 7">
      </path>
     </svg>
    </router-link>
    <router-link :to="pricingPath" class="group/button inline-flex shrink-0 items-center justify-center border bg-clip-padding text-sm font-medium whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&amp;_svg]:pointer-events-none [&amp;_svg]:shrink-0 [&amp;_svg:not([class*='size-'])]:size-4 bg-background hover:text-foreground aria-expanded:bg-muted aria-expanded:text-foreground dark:border-input dark:bg-input/30 dark:hover:bg-input/50 h-8 gap-1.5 px-2.5 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2 border-border/50 hover:border-border hover:bg-muted/50 rounded-lg" data-slot="button" tabindex="0">
     {{ l('查看定价') }}
    </router-link>
   </div>
  </div>
 </section>
 <footer class="border-border/40 relative z-10 border-t">
  <div class="mx-auto max-w-6xl px-6 py-12 md:py-16">
   <div class="flex flex-col justify-between gap-10 md:flex-row md:gap-16">
    <div class="shrink-0">
     <router-link class="group flex items-center gap-2.5 active" to="/home">
      <img :alt="siteName" :src="siteLogo" class="size-7 rounded-lg object-contain"/>
      <span class="text-sm font-semibold tracking-tight">
       {{ siteName }}
      </span>
     </router-link>
     <p class="text-muted-foreground/60 mt-3 max-w-[200px] text-xs leading-relaxed">
      {{ l('强大的 API 管理平台') }}
     </p>
    </div>
   </div>
   <div class="border-border/30 mt-12 flex flex-col items-center justify-between gap-x-3 gap-y-2 border-t pt-6 sm:flex-row">
    <div class="text-muted-foreground/40 flex flex-wrap items-center justify-center gap-x-2 gap-y-1 text-xs sm:justify-start">
     <span>
      © {{ currentYear }} {{ siteName }}. {{ l('版权所有。') }}
     </span>
    </div>
    <div class="text-muted-foreground/45 text-center text-xs sm:text-right">
     <span class="text-muted-foreground/45">
      {{ l('由{siteName}提供服务') }}
     </span>
    </div>
   </div>
  </div>
 </footer>
</div>

</template>

<script setup lang="ts">

import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import { availableLocales, getLocale, setLocale } from '@/i18n'
import { landingProtocols } from './landing-protocols'
import { translateLandingCopy, type LandingCopyKey } from './landing-copy'
import './wugou-landing.css'

const props = defineProps<{
  siteName: string
  siteLogo: string
  docUrl: string
  isAuthenticated: boolean
  dashboardPath: string
  showModelPlazaEntry: boolean
  registrationEnabled: boolean
  isDark: boolean
}>()
const emit = defineEmits<{ 'toggle-theme': [] }>()
const currentYear = new Date().getFullYear()
const documentationUrl = computed(() => props.docUrl || '/workbuddy')
const startPath = computed(() => props.isAuthenticated ? props.dashboardPath : props.registrationEnabled ? '/register' : '/login')
const pricingPath = computed(() => props.showModelPlazaEntry ? '/model-plaza' : '/subscriptions')
const scrolled = ref(false)
const mobileMenuOpen = ref(false)
const mobileMenu = ref<HTMLElement>()
const mobileMenuButton = ref<HTMLButtonElement>()
const languageMenu = ref<HTMLElement>()
const languageMenuOpen = ref(false)
const currentLocale = ref(getLocale())
const l = (key: LandingCopyKey) => translateLandingCopy(key, currentLocale.value, props.siteName)
const activeProtocol = ref(0)
let rotation: ReturnType<typeof setInterval> | undefined
let previousOverflow = ''

function stopRotation() {
  if (rotation) clearInterval(rotation)
  rotation = undefined
}
function selectProtocol(index: number) {
  stopRotation()
  activeProtocol.value = index
}
function onProtocolKeydown(event: KeyboardEvent) {
  let index = activeProtocol.value
  if (event.key === 'ArrowRight') index = (index + 1) % landingProtocols.length
  else if (event.key === 'ArrowLeft') index = (index + landingProtocols.length - 1) % landingProtocols.length
  else if (event.key === 'Home') index = 0
  else if (event.key === 'End') index = landingProtocols.length - 1
  else return
  event.preventDefault()
  selectProtocol(index)
  nextTick(() => document.getElementById('protocol-tab-' + landingProtocols[index].id)?.focus())
}
async function changeLanguage(code: string) {
  await setLocale(code)
  currentLocale.value = getLocale()
  languageMenuOpen.value = false
}
function onScroll() { scrolled.value = window.scrollY > 40 }
function onDocumentClick(event: MouseEvent) {
  if (!languageMenu.value?.contains(event.target as Node)) languageMenuOpen.value = false
}
function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    languageMenuOpen.value = false
    mobileMenuOpen.value = false
  }
  if (event.key !== 'Tab' || !mobileMenuOpen.value) return
  const items = [mobileMenuButton.value, ...Array.from(mobileMenu.value?.querySelectorAll<HTMLElement>('a, button') || [])].filter((e): e is HTMLElement => !!e)
  const first = items[0], last = items[items.length - 1]
  if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last?.focus() }
  else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first?.focus() }
}
function onResize() {
  if (window.innerWidth >= 1024) mobileMenuOpen.value = false
}
watch(mobileMenuOpen, async (open) => {
  if (open) {
    previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    await nextTick()
    mobileMenu.value?.querySelector<HTMLElement>('a')?.focus()
  } else {
    document.body.style.overflow = previousOverflow
    mobileMenuButton.value?.focus()
  }
})
onMounted(() => {
  onScroll()
  window.addEventListener('scroll', onScroll, { passive: true })
  window.addEventListener('resize', onResize)
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onKeydown)
  if (!window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
    rotation = setInterval(() => {
      if (!document.hidden && !document.getElementById('protocol-panel')?.contains(document.activeElement) && !document.activeElement?.closest('[role="tablist"]')) {
        activeProtocol.value = (activeProtocol.value + 1) % landingProtocols.length
      }
    }, 4500)
  }
})
onBeforeUnmount(() => {
  stopRotation()
  if (mobileMenuOpen.value) document.body.style.overflow = previousOverflow
  window.removeEventListener('scroll', onScroll)
  window.removeEventListener('resize', onResize)
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onKeydown)
})
</script>

<style scoped>
.landing-locale { position: relative; display: flex; }
.landing-language-menu { position: absolute; top: calc(100% + 8px); right: 0; width: 148px; padding: 4px; background: var(--background); border: 1px solid var(--border); border-radius: 12px; box-shadow: 0 8px 28px rgb(0 0 0 / 10%); }
.landing-language-menu button { display: flex; justify-content: space-between; width: 100%; padding: 8px 12px; border-radius: 8px; font-size: 14px; text-align: left; }
.landing-language-menu button:hover { background: var(--muted); }
.landing-protocol-tabs { display: flex; min-width: 0; gap: 4px; }
@media (min-width: 640px) { .landing-protocol-tabs { gap: 6px; } }
@media (prefers-reduced-motion: reduce) { .wugou-landing :deep(*) { animation: none !important; transition: none !important; } }
</style>
