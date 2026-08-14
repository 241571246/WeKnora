<template>
  <div class="vone-brand-identity" :class="[`vone-brand-identity--${size}`, { 'is-mark-only': !showName }]">
    <picture class="vone-brand-identity__logo-wrap">
      <img class="vone-brand-identity__logo vone-brand-identity__logo--blue" :src="voneBrand.logos.blue"
        :alt="voneBrand.shortName">
      <img class="vone-brand-identity__logo vone-brand-identity__logo--light" :src="voneBrand.logos.light"
        :alt="voneBrand.shortName">
    </picture>
    <span v-if="showName" class="vone-brand-identity__names">
      <strong>{{ voneBrand.productNameZh }}</strong>
      <small v-if="showEnglish">{{ voneBrand.productNameEn }}</small>
    </span>
  </div>
</template>

<script setup lang="ts">
import { voneBrand } from '../brand';

withDefaults(defineProps<{
  showName?: boolean;
  showEnglish?: boolean;
  size?: 'compact' | 'default' | 'hero';
}>(), {
  showName: true,
  showEnglish: true,
  size: 'default',
});
</script>

<style scoped>
.vone-brand-identity {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
  color: var(--vone-text-primary);
}

.vone-brand-identity__logo-wrap {
  position: relative;
  display: block;
  width: 154px;
  flex: 0 0 auto;
  line-height: 0;
}

.vone-brand-identity__logo {
  display: block;
  width: 100%;
  height: auto;
  object-fit: contain;
}

.vone-brand-identity__logo--light {
  position: absolute;
  inset: 0;
  opacity: var(--vone-logo-light-opacity, 0);
}

.vone-brand-identity__logo--blue {
  opacity: var(--vone-logo-blue-opacity, 1);
}

.vone-brand-identity__names {
  display: flex;
  min-width: 0;
  flex-direction: column;
  line-height: 1.15;
}

.vone-brand-identity__names strong,
.vone-brand-identity__names small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.vone-brand-identity__names strong {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.vone-brand-identity__names small {
  margin-top: 4px;
  color: var(--vone-text-secondary);
  font-size: 12px;
}

.vone-brand-identity--compact {
  gap: 9px;
}

.vone-brand-identity--compact .vone-brand-identity__logo-wrap {
  width: 58px;
}

.vone-brand-identity--compact .vone-brand-identity__names strong {
  font-size: 14px;
}

.vone-brand-identity--compact .vone-brand-identity__names small {
  display: none;
}

.vone-brand-identity--hero {
  align-items: flex-start;
  flex-direction: column;
  gap: 28px;
}

.vone-brand-identity--hero .vone-brand-identity__logo-wrap {
  width: 202px;
}

.vone-brand-identity--hero .vone-brand-identity__names strong {
  font-size: clamp(38px, 4vw, 58px);
}

.vone-brand-identity--hero .vone-brand-identity__names small {
  margin-top: 8px;
  font-size: clamp(20px, 2vw, 28px);
}

</style>
