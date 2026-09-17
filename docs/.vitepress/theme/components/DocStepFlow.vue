<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { Swiper, SwiperSlide } from 'swiper/vue'
import { Autoplay, EffectFade, Keyboard } from 'swiper/modules'
import type { Swiper as SwiperType } from 'swiper'

import 'swiper/css'
import 'swiper/css/effect-fade'
import '@fancyapps/ui/dist/fancybox/fancybox.css'

export interface IStepItem {
  id?: string | number
  title: string
  tag?: string
  url?: string
  image: string
}

export interface IDocStepFlowProps {
  steps: IStepItem[]
  autoplay?: boolean
  interval?: number
}

// NOTE: 遵循单向数据流规范声明入参
const props = withDefaults(defineProps<IDocStepFlowProps>(), {
  autoplay: true,
  interval: 6000
})

const swiperInstance = ref<SwiperType | null>(null)
const currentIndex = ref(0)
const isPlaying = ref(props.autoplay)

const currentStep = computed(() => props.steps?.[currentIndex.value] ?? null)

const onSwiperInit = (swiper: SwiperType) => {
  swiperInstance.value = swiper
}

const onSlideChange = (swiper: SwiperType) => {
  currentIndex.value = swiper.realIndex
}

const selectStep = (idx: number) => {
  swiperInstance.value?.slideTo(idx)
}

const copied = ref(false)
const copyUrl = async () => {
  if (!currentStep.value?.url) return
  try {
    await navigator.clipboard.writeText(currentStep.value.url)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 1800)
  } catch {
    // 降级支持
  }
}

const togglePlay = () => {
  if (!swiperInstance.value) return
  if (isPlaying.value) {
    swiperInstance.value.autoplay.stop()
    isPlaying.value = false
  } else {
    swiperInstance.value.autoplay.start()
    isPlaying.value = true
  }
}

onMounted(async () => {
  // 客户端动态按需加载专业画廊库 Fancybox，接管所有步骤的全屏高清放大与连续翻页
  const { Fancybox } = await import('@fancyapps/ui')
  let savedScrollY = 0

  Fancybox.bind('[data-fancybox="doc-steps"]', {
    // 禁用 Hash 插件，杜绝其在关闭时执行 history.back() 触发 VitePress 路由重置滚回顶部
    Hash: false,
    // 禁用关闭时强制对焦 triggerEl 避免触发浏览器的 scrollIntoView 导致画面跳变
    placeFocusBack: false,
    Carousel: {
      Thumbs: {
        showOnStart: false
      },
      Toolbar: {
        display: {
          left: ['counter'],
          middle: ['zoomIn', 'zoomOut', 'toggle1to1'],
          right: ['autoplay', 'thumbs', 'close']
        }
      }
    },
    on: {
      init: () => {
        // 记录打开画廊前的精确滚动位置
        savedScrollY = window.scrollY
        swiperInstance.value?.autoplay.stop()
      },
      close: () => {
        if (isPlaying.value) {
          swiperInstance.value?.autoplay.start()
        }
        // 双重保障：确保画廊关闭后页面坚挺停留在原阅读位置
        requestAnimationFrame(() => {
          window.scrollTo({ top: savedScrollY, behavior: 'instant' })
        })
      },
      'Carousel.change': (_fancybox: any, _carousel: any, to: number) => {
        swiperInstance.value?.slideTo(to)
      }
    }
  })
})

onUnmounted(async () => {
  const { Fancybox } = await import('@fancyapps/ui')
  Fancybox.destroy()
})
</script>

<template>
  <div class="macos-walkthrough">
    <!-- 1. macOS 拟真窗口工具栏 -->
    <div class="window-chrome">
      <div class="chrome-dots">
        <span class="dot close"></span>
        <span class="dot minimize"></span>
        <span class="dot maximize"></span>
      </div>

      <!-- 拟真 URL 地址胶囊（支持鼠标自由划选与一键复制按钮） -->
      <div class="chrome-url-bar">
        <svg class="url-icon" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
          <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
        </svg>
        <span class="url-text" :title="currentStep?.url">{{ currentStep?.url || 'fleetops.top/cmdb/plugins' }}</span>
        
        <button 
          type="button" 
          class="url-copy-btn" 
          :title="copied ? '已复制到剪贴板' : '一键复制路由'"
          @click.stop="copyUrl"
        >
          <span v-if="copied" class="copied-icon">✓</span>
          <svg v-else class="copy-svg" viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
          </svg>
        </button>

        <span v-if="currentStep?.tag" class="url-badge">{{ currentStep.tag }}</span>
      </div>

      <div class="chrome-actions">
        <button 
          type="button" 
          class="ctrl-icon-btn" 
          @click="togglePlay" 
          :title="isPlaying ? '暂停自动轮播' : '恢复自动播放'"
        >
          <span v-if="isPlaying" class="icon-pause">⏸</span>
          <span v-else class="icon-play">▶</span>
        </button>

        <span class="step-digits">{{ currentIndex + 1 }}/{{ props.steps.length }}</span>

        <div class="nav-arrows">
          <button type="button" class="arrow-btn" @click="swiperInstance?.slidePrev()" title="上一个">‹</button>
          <button type="button" class="arrow-btn" @click="swiperInstance?.slideNext()" title="下一个">›</button>
        </div>
      </div>
    </div>

    <!-- 2. 基于 Swiper 的视窗 + 基于 Fancybox 的全屏专业画廊（100% 纯粹大图查看） -->
    <div class="window-viewport">
      <ClientOnly>
        <Swiper
          :modules="[Autoplay, EffectFade, Keyboard]"
          :effect="'fade'"
          :fade-effect="{ crossFade: true }"
          :autoplay="props.autoplay ? { delay: props.interval, pauseOnMouseEnter: true, disableOnInteraction: false } : false"
          :speed="400"
          :loop="false"
          :grab-cursor="true"
          @swiper="onSwiperInit"
          @slide-change="onSlideChange"
        >
          <SwiperSlide v-for="(step, idx) in props.steps" :key="idx">
            <a 
              :href="step.image" 
              data-fancybox="doc-steps"
              class="screen-link no-zoom"
              title="点击全屏高清放大（支持滚轮缩放与快捷切图）"
            >
              <img :src="step.image" :alt="step.title" class="screen-image no-zoom" />
            </a>
          </SwiperSlide>
        </Swiper>
      </ClientOnly>

      <!-- 悬停指示 -->
      <div class="zoom-hint">
        <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"></circle>
          <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          <line x1="11" y1="8" x2="11" y2="14"></line>
          <line x1="8" y1="11" x2="14" y2="11"></line>
        </svg>
        <span>点击全屏放大</span>
      </div>
    </div>

    <!-- 3. 步骤分段选择条（极简现代胶囊） -->
    <div class="steps-segmented-bar">
      <button 
        v-for="(step, idx) in props.steps" 
        :key="idx"
        type="button"
        class="segmented-item"
        :class="{ 'is-active': currentIndex === idx }"
        @click="selectStep(idx)"
      >
        <span class="seg-badge">{{ idx + 1 }}</span>
        <span class="seg-title">{{ step.title }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.macos-walkthrough {
  margin: 28px 0 20px;
  background: var(--vp-c-bg-elv);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 12px 32px -8px rgba(0, 0, 0, 0.14);
  transition: all 0.25s ease;
}

.macos-walkthrough:hover {
  border-color: var(--vp-c-brand-1);
}

/* 顶部 macOS 窗口工具栏 */
.window-chrome {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
  background: var(--vp-c-bg-alt);
  border-bottom: 1px solid var(--vp-c-divider);
  gap: 12px;
}

.chrome-dots {
  display: flex;
  gap: 6px;
  width: 54px;
  user-select: none;
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
}

.dot.close { background: #ff5f56; }
.dot.minimize { background: #ffbd2e; }
.dot.maximize { background: #27c93f; }

.chrome-url-bar {
  flex: 1;
  max-width: 480px;
  display: flex;
  align-items: center;
  gap: 6px;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  padding: 3px 8px;
  font-family: var(--vp-font-family-mono);
  font-size: 11px;
  color: var(--vp-c-text-2);
  user-select: text;
}

.url-icon { color: #10b981; flex-shrink: 0; }

.url-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--vp-c-text-1);
  user-select: text !important;
  cursor: text;
}

.url-copy-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: var(--vp-c-text-3);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 4px;
  flex-shrink: 0;
  transition: all 0.15s ease;
}

.url-copy-btn:hover {
  color: var(--vp-c-brand-1);
  background: var(--vp-c-brand-soft);
}

.copied-icon {
  font-size: 11px;
  color: #10b981;
  font-weight: 700;
  line-height: 1;
}

.copy-svg {
  display: block;
}

.url-badge {
  flex-shrink: 0 !important;
  white-space: nowrap !important;
  margin-left: 2px;
  font-size: 10px;
  line-height: 1.4;
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand-1);
  font-weight: 600;
}

.chrome-actions {
  display: flex;
  align-items: center;
  gap: 7px;
  flex-shrink: 0 !important;
  white-space: nowrap !important;
}

.ctrl-icon-btn {
  width: 24px;
  height: 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 5px;
  font-size: 11px;
  color: var(--vp-c-text-2);
  cursor: pointer;
  flex-shrink: 0 !important;
  transition: all 0.15s ease;
}

.ctrl-icon-btn:hover {
  color: var(--vp-c-brand-1);
  border-color: var(--vp-c-brand-1);
  background: var(--vp-c-brand-soft);
}

.icon-pause, .icon-play {
  font-size: 10px;
  line-height: 1;
}

.step-digits {
  font-family: var(--vp-font-family-mono);
  font-size: 11px;
  font-weight: 600;
  color: var(--vp-c-text-3);
  flex-shrink: 0 !important;
  white-space: nowrap !important;
  padding: 0 2px;
}

.nav-arrows {
  display: flex;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 5px;
  flex-shrink: 0 !important;
  white-space: nowrap !important;
}

.arrow-btn {
  padding: 2px 7px;
  font-size: 14px;
  line-height: 1;
  color: var(--vp-c-text-2);
  background: transparent;
  border: none;
  cursor: pointer;
}

.arrow-btn:hover { background: var(--vp-c-brand-soft); color: var(--vp-c-brand-1); }

/* 主视窗 */
.window-viewport {
  position: relative;
  background: #000000;
  overflow: hidden;
}

.screen-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  cursor: zoom-in;
  text-decoration: none !important;
}

.screen-image {
  width: 100%;
  height: auto;
  display: block;
  object-fit: contain;
  margin: 0 !important;
  border-radius: 0 !important;
}

/* 悬停微提示 */
.zoom-hint {
  position: absolute;
  right: 12px;
  bottom: 12px;
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 8px;
  background: rgba(0, 0, 0, 0.72);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 6px;
  color: #d1d5db;
  font-size: 10.5px;
  pointer-events: none;
  opacity: 0;
  transform: translateY(4px);
  transition: all 0.25s ease;
  z-index: 10;
}

.macos-walkthrough:hover .zoom-hint {
  opacity: 1;
  transform: translateY(0);
}

/* 分段步骤导航条 */
.steps-segmented-bar {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  background: var(--vp-c-bg-alt);
  border-top: 1px solid var(--vp-c-divider);
  padding: 5px 6px;
  gap: 4px;
}

.segmented-item {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding: 6px 3px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.segmented-item:hover { background: var(--vp-c-bg-soft); }

.segmented-item.is-active {
  background: var(--vp-c-bg-elv);
  border-color: var(--vp-c-divider);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.seg-badge {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--vp-c-divider);
  color: var(--vp-c-text-2);
  font-size: 10px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.segmented-item.is-active .seg-badge {
  background: var(--vp-c-brand-1);
  color: #ffffff;
}

.seg-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--vp-c-text-2);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.segmented-item.is-active .seg-title { color: var(--vp-c-brand-1); }

@media (max-width: 768px) {
  .chrome-url-bar { display: none; }
  .steps-segmented-bar { grid-template-columns: repeat(3, 1fr); }
}
</style>
