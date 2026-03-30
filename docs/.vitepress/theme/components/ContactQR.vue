<script setup lang="ts">
import { ref } from 'vue'

// NOTE: 该组件仅用于在导航栏展示“交流讨论”并支持悬停显示二维码
const showQR = ref(false)
</script>

<template>
  <div class="contact-qr-container" @mouseenter="showQR = true" @mouseleave="showQR = false">
    <div class="contact-trigger">
      <span class="text">交流讨论</span>
      <span class="icon">💬</span>
    </div>
    
    <Transition name="fade">
      <div v-if="showQR" class="qr-popover">
        <div class="qr-card">
          <p class="qr-title">扫码添加微信</p>
          <img src="/wechat-qr.jpg" alt="WeChat QR" class="qr-image" />
          <p class="qr-desc">备注：ECMDB</p>
        </div>
        <div class="qr-arrow"></div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.contact-qr-container {
  position: relative;
  display: flex;
  align-items: center;
  padding: 0 12px;
  height: var(--vp-nav-height);
  cursor: pointer;
}

.contact-trigger {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  font-weight: 500;
  color: var(--vp-c-text-1);
  transition: color 0.25s;
}

.contact-qr-container:hover .contact-trigger {
  color: var(--vp-c-brand-1);
}

.text {
  line-height: var(--vp-nav-height);
}

.icon {
  font-size: 16px;
}

.qr-popover {
  position: absolute;
  top: calc(var(--vp-nav-height) - 10px);
  right: 0;
  padding-top: 15px;
  z-index: 100;
  width: 220px;
}

.qr-card {
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  padding: 16px;
  box-shadow: var(--vp-shadow-3);
  text-align: center;
}

.qr-title {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
  color: var(--vp-c-brand-1);
}

.qr-image {
  width: 100%;
  aspect-ratio: 1;
  border-radius: 8px;
  border: 1px solid var(--vp-c-divider-light);
}

.qr-desc {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--vp-c-text-2);
}

.qr-arrow {
  position: absolute;
  top: 7px;
  right: 24px;
  width: 16px;
  height: 16px;
  background: var(--vp-c-bg);
  border-left: 1px solid var(--vp-c-divider);
  border-top: 1px solid var(--vp-c-divider);
  transform: rotate(45deg);
}

/* 动画效果 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(5px);
}

/* 适配移动端：隐藏或特殊处理 (VitePress 默认有 nav-screen) */
@media (max-width: 959px) {
  .contact-qr-container {
    display: none;
  }
}
</style>
