<template>
  <div class="vone-system-switcher" :class="{ 'vone-system-switcher--compact': compact }">
    <t-dropdown
      trigger="click"
      :placement="compact ? 'right-top' : 'bottom-left'"
      attach="body"
      :min-column-width="228"
    >
      <button
        v-if="compact"
        type="button"
        class="system-switcher-compact"
        aria-label="切换系统"
        title="切换系统"
      >
        <svg viewBox="0 0 20 20" width="20" height="20" fill="none" aria-hidden="true">
          <rect x="2" y="2" width="6.5" height="6.5" rx="1.5" stroke="currentColor" stroke-width="1.3" />
          <rect x="11.5" y="2" width="6.5" height="6.5" rx="1.5" stroke="currentColor" stroke-width="1.3" />
          <rect x="2" y="11.5" width="6.5" height="6.5" rx="1.5" stroke="currentColor" stroke-width="1.3" />
          <rect x="11.5" y="11.5" width="6.5" height="6.5" rx="1.5" stroke="currentColor" stroke-width="1.3" />
        </svg>
      </button>
      <button v-else type="button" class="system-switcher-trigger" aria-label="切换系统">
        <span class="logo-box">
          <BrandIdentity size="compact" :show-english="false" />
          <sup v-if="liteEdition" class="lite-badge">Lite</sup>
        </span>
        <svg class="system-switcher-chevron" viewBox="0 0 16 16" width="14" height="14" fill="none"
          aria-hidden="true">
          <path d="m4 6 4 4 4-4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"
            stroke-linejoin="round" />
        </svg>
      </button>

      <template #dropdown>
        <t-dropdown-menu class="system-switcher-menu">
          <t-dropdown-item disabled>
            <span class="system-switcher-menu__icon system-switcher-menu__icon--current">
              <svg viewBox="0 0 16 16" width="15" height="15" fill="none" aria-hidden="true">
                <path d="m3.5 8 3 3 6-6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"
                  stroke-linejoin="round" />
              </svg>
            </span>
            <span class="system-switcher-menu__label">AI大脑平台</span>
            <span class="system-switcher-menu__current">当前</span>
          </t-dropdown-item>
          <t-dropdown-item v-if="targetUrl" @click="openTargetSystem">
            <span class="system-switcher-menu__icon">
              <svg viewBox="0 0 16 16" width="15" height="15" fill="none" aria-hidden="true">
                <rect x="2" y="2.5" width="12" height="11" rx="2" stroke="currentColor" stroke-width="1.2" />
                <path d="M5 6h6M5 8.5h4M5 11h5" stroke="currentColor" stroke-width="1.2"
                  stroke-linecap="round" />
              </svg>
            </span>
            <span class="system-switcher-menu__label">项目管理系统</span>
            <svg class="system-switcher-menu__external" viewBox="0 0 16 16" width="14" height="14" fill="none"
              aria-hidden="true">
              <path d="M6 3h7v7M13 3 6.5 9.5M11 8.5V12a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1h3.5"
                stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </t-dropdown-item>
        </t-dropdown-menu>
      </template>
    </t-dropdown>
  </div>
</template>

<script setup lang="ts">
import BrandIdentity from './BrandIdentity.vue';

const props = withDefaults(defineProps<{
  targetUrl: string | null;
  compact?: boolean;
  liteEdition?: boolean;
}>(), {
  compact: false,
  liteEdition: false,
});

function openTargetSystem() {
  if (!props.targetUrl) return;
  const openedWindow = window.open(props.targetUrl, '_blank', 'noopener,noreferrer');
  if (openedWindow) openedWindow.opener = null;
}
</script>

<style scoped lang="less">
.vone-system-switcher {
  display: flex;
  flex: 1;
  min-width: 0;
}

.vone-system-switcher--compact {
  flex: 0 0 auto;
  justify-content: center;
  width: 100%;
  margin: 2px 0 8px;
}

.system-switcher-trigger {
  display: flex;
  align-items: center;
  width: 100%;
  min-width: 0;
  height: 38px;
  margin-left: -6px;
  padding: 0 6px;
  border: 0;
  border-radius: 7px;
  color: var(--td-text-color-primary);
  background: transparent;
  cursor: pointer;
  transition: background-color 0.2s ease;

  &:hover,
  &:focus-visible {
    background: var(--td-bg-color-container-hover);
    outline: none;
  }
}

.logo-box {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  overflow: hidden;

  :deep(.vone-brand-identity),
  :deep(.vone-brand-identity__names) {
    min-width: 0;
  }

  :deep(.vone-brand-identity__names) {
    max-width: 116px;
  }
}

.lite-badge {
  margin-top: 2px;
  margin-left: 2px;
  align-self: flex-start;
  color: var(--td-text-color-placeholder);
  font-size: 9px;
  font-weight: 600;
  white-space: nowrap;
  user-select: none;
}

.system-switcher-chevron {
  flex-shrink: 0;
  margin-left: 4px;
  color: var(--td-text-color-placeholder);
}

.system-switcher-compact {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  padding: 0;
  border: 0;
  border-radius: 6px;
  color: var(--td-text-color-secondary);
  background: transparent;
  cursor: pointer;
  transition: color 0.2s ease, background-color 0.2s ease;

  &:hover,
  &:focus-visible {
    color: var(--td-brand-color);
    background: var(--td-brand-color-light);
    outline: none;
  }
}

.system-switcher-menu__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 18px;
  color: var(--td-text-color-secondary);

  &--current {
    color: var(--td-brand-color);
  }
}

.system-switcher-menu__label {
  flex: 1;
  min-width: 0;
  margin-left: 8px;
  color: var(--td-text-color-primary);
  font-size: 13px;
  font-weight: 500;
}

.system-switcher-menu__current {
  margin-left: 14px;
  padding: 1px 6px;
  border-radius: 999px;
  color: var(--td-brand-color);
  background: var(--td-brand-color-light);
  font-size: 11px;
  line-height: 18px;
}

.system-switcher-menu__external {
  flex-shrink: 0;
  margin-left: 14px;
  color: var(--td-text-color-placeholder);
}
</style>
