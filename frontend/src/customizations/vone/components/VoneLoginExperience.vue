<template>
  <main class="vone-auth-shell">
    <section class="vone-auth-story" aria-label="Enterprise AI Brain Platform">
      <div class="vone-auth-story__glow vone-auth-story__glow--top" />
      <div class="vone-auth-story__glow vone-auth-story__glow--bottom" />

      <BrandIdentity size="hero" />

      <div class="vone-auth-story__copy">
        <p>{{ copy.storyLineOne }}</p>
        <p>{{ copy.storyLineTwo }}</p>
      </div>

      <div class="knowledge-visual" aria-hidden="true">
        <svg class="knowledge-visual__links" viewBox="0 0 720 420" preserveAspectRatio="none">
          <path d="M42 278 L146 214 L242 250 L332 164 L454 224 L566 130 L688 182" />
          <path d="M74 332 L176 300 L286 348 L396 274 L512 322 L654 260" />
          <path d="M146 214 L176 300 M242 250 L286 348 M332 164 L396 274 M454 224 L512 322 M566 130 L654 260" />
        </svg>
        <div class="knowledge-node knowledge-node--one"><span>W</span></div>
        <div class="knowledge-node knowledge-node--two"><span>⌕</span></div>
        <div class="knowledge-node knowledge-node--three"><span>▤</span></div>
        <div class="knowledge-node knowledge-node--four"><span>◫</span></div>
        <div class="knowledge-node knowledge-node--five"><span>◉</span></div>

        <div class="knowledge-card knowledge-card--document">
          <span class="knowledge-card__icon">W</span>
          <i /><i /><i />
        </div>
        <div class="knowledge-card knowledge-card--answer">
          <span class="knowledge-card__icon">AI</span>
          <i /><i /><i />
        </div>
        <div class="knowledge-card knowledge-card--data">
          <span class="knowledge-card__icon">▦</span>
          <i /><i /><i />
        </div>
        <div class="knowledge-search">
          <span class="knowledge-search__icon">⌕</span>
          <span>{{ copy.searchPlaceholder }}</span>
        </div>
      </div>
    </section>

    <section class="vone-auth-workspace">
      <header class="vone-auth-toolbar">
        <div class="theme-segment" :aria-label="copy.themeLabel">
          <button v-for="option in themeOptions" :key="option.value" type="button"
            :class="{ active: currentTheme === option.value }" @click="setTheme(option.value)">
            <span aria-hidden="true">{{ option.icon }}</span>
            {{ option.label }}
          </button>
        </div>

        <div class="language-switch">
          <button type="button" class="language-trigger" :aria-expanded="showLanguageMenu"
            @click.stop="$emit('toggle-language')">
            <span aria-hidden="true">◎</span>
            {{ currentLangOption?.shortLabel || '中文' }}
            <span class="language-trigger__chevron">⌄</span>
          </button>
          <div v-if="showLanguageMenu" class="language-dropdown">
            <button v-for="lang in languageOptions" :key="lang.value" type="button"
              :class="{ active: currentLanguage === lang.value }" @click="$emit('select-language', lang.value)">
              <span>{{ lang.flag }}</span>
              <span>{{ lang.label }}</span>
              <span v-if="currentLanguage === lang.value" class="language-check">✓</span>
            </button>
          </div>
        </div>
      </header>

      <div class="vone-auth-panel">
        <div v-if="!isRegisterMode" class="vone-auth-card">
          <div class="vone-auth-card__header">
            <span class="vone-auth-card__eyebrow">{{ voneBrand.productNameEn }}</span>
            <h1>{{ copy.loginTitle }}</h1>
            <p>{{ copy.loginSubtitle }}</p>
          </div>

          <t-form ref="loginFormRef" :data="formData" :rules="formRules" layout="vertical" label-align="top"
            @submit="$emit('login')">
            <t-form-item :label="$t('auth.email')" name="email">
              <t-input v-model="formData.email" :placeholder="$t('auth.emailPlaceholder')" type="text"
                autocomplete="email" size="large" :disabled="loading">
                <template #prefix-icon><span class="field-icon">♙</span></template>
              </t-input>
            </t-form-item>

            <t-form-item :label="$t('auth.password')" name="password">
              <t-input v-model="formData.password" :placeholder="$t('auth.passwordPlaceholder')" type="password"
                autocomplete="current-password" size="large" :disabled="loading" @enter="$emit('login')">
                <template #prefix-icon><span class="field-icon">▣</span></template>
              </t-input>
            </t-form-item>

            <t-button type="submit" theme="primary" size="large" block :loading="loading" class="primary-action">
              {{ loading ? $t('auth.loggingIn') : $t('auth.login') }}
            </t-button>

            <t-button v-if="oidcEnabled" theme="default" variant="outline" size="large" block
              :loading="oidcLoading" :disabled="loading" class="secondary-action" @click="$emit('oidc-login')">
              <span class="enterprise-icon">▥</span>
              {{ oidcLoading ? $t('auth.redirectingToOIDC') : oidcLoginText }}
            </t-button>
          </t-form>

          <div v-if="registrationEnabled" class="auth-mode-link">
            <span>{{ copy.noAccount }}</span>
            <button type="button" :disabled="loading" @click="$emit('toggle-mode')">{{ $t('auth.createAccount') }}</button>
          </div>
        </div>

        <div v-else-if="registrationEnabled || inviteLookup" class="vone-auth-card vone-auth-card--register">
          <div v-if="inviteLookup" class="invite-banner">
            <span class="invite-banner__icon">↗</span>
            <div>
              <strong>{{ $t('inviteRegister.bannerTitle', { tenant: inviteLookup.tenant_name || '' }) }}</strong>
              <p>{{ $t('inviteRegister.bannerHint') }}</p>
            </div>
          </div>
          <div v-else-if="inviteLookupError" class="invite-banner invite-banner--error">{{ inviteLookupError }}</div>

          <div class="vone-auth-card__header">
            <span class="vone-auth-card__eyebrow">{{ voneBrand.productNameEn }}</span>
            <h1>{{ $t('auth.createAccount') }}</h1>
            <p>{{ copy.registerSubtitle }}</p>
          </div>

          <t-form ref="registerFormRef" :data="registerData" :rules="registerRules" layout="vertical"
            label-align="top" @submit="$emit('register')">
            <div class="register-grid">
              <t-form-item :label="$t('auth.username')" name="username">
                <t-input v-model="registerData.username" :placeholder="$t('auth.usernamePlaceholder')" size="large"
                  :disabled="loading" />
              </t-form-item>
              <t-form-item :label="$t('auth.email')" name="email">
                <t-input v-model="registerData.email" :placeholder="$t('auth.emailPlaceholder')" type="text"
                  autocomplete="email" size="large" :disabled="loading" />
              </t-form-item>
              <t-form-item :label="$t('auth.password')" name="password">
                <t-input v-model="registerData.password" :placeholder="$t('auth.passwordPlaceholder')" type="password"
                  autocomplete="new-password" size="large" :disabled="loading" />
              </t-form-item>
              <t-form-item :label="$t('auth.confirmPassword')" name="confirmPassword">
                <t-input v-model="registerData.confirmPassword" :placeholder="$t('auth.confirmPasswordPlaceholder')"
                  type="password" autocomplete="new-password" size="large" :disabled="loading"
                  @enter="$emit('register')" />
              </t-form-item>
            </div>
            <t-button type="submit" theme="primary" size="large" block :loading="loading" class="primary-action">
              {{ loading ? $t('auth.registering') : $t('auth.register') }}
            </t-button>
          </t-form>

          <div class="auth-mode-link">
            <span>{{ $t('auth.haveAccount') }}</span>
            <button type="button" :disabled="loading" @click="$emit('toggle-mode')">{{ $t('auth.backToLogin') }}</button>
          </div>
        </div>
      </div>
    </section>

    <footer class="vone-auth-footer">
      <span>{{ voneBrand.copyright }}</span>
      <a :href="voneBrand.upstreamUrl" target="_blank" rel="noreferrer">{{ voneBrand.upstreamAttribution }}</a>
    </footer>
  </main>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useTheme, type ThemeMode } from '@/composables/useTheme';
import { voneBrand } from '../brand';
import BrandIdentity from './BrandIdentity.vue';

type LanguageOption = {
  value: string;
  label: string;
  shortLabel: string;
  flag: string;
};

const props = defineProps<{
  isRegisterMode: boolean;
  registrationEnabled: boolean;
  inviteLookup: Record<string, any> | null;
  inviteLookupError: string;
  loading: boolean;
  oidcEnabled: boolean;
  oidcLoading: boolean;
  oidcLoginText: string;
  showLanguageMenu: boolean;
  languageOptions: LanguageOption[];
  currentLanguage: string;
  currentLangOption?: LanguageOption;
  formData: Record<string, any>;
  registerData: Record<string, any>;
  formRules: Record<string, any>;
  registerRules: Record<string, any>;
}>();

defineEmits<{
  (event: 'login'): void;
  (event: 'register'): void;
  (event: 'oidc-login'): void;
  (event: 'toggle-mode'): void;
  (event: 'toggle-language'): void;
  (event: 'select-language', language: string): void;
}>();

const loginFormRef = ref();
const registerFormRef = ref();
const { locale } = useI18n();
const { currentTheme, setTheme } = useTheme();

const isChinese = computed(() => locale.value === 'zh-CN');
const copy = computed(() => isChinese.value ? {
  storyLineOne: '连接知识，激活企业智慧',
  storyLineTwo: '让每一份知识都产生价值',
  searchPlaceholder: '搜索知识、文档、问题或答案…',
  loginTitle: '欢迎登录',
  loginSubtitle: '连接知识、检索答案，构建可信的企业智能',
  registerSubtitle: '创建账户并开始使用企业AI大脑平台',
  noAccount: '还没有账户？',
  themeLabel: '界面主题',
} : {
  storyLineOne: 'Connect knowledge. Activate enterprise intelligence.',
  storyLineTwo: 'Make every piece of knowledge valuable.',
  searchPlaceholder: 'Search knowledge, documents, questions or answers…',
  loginTitle: 'Welcome back',
  loginSubtitle: 'Connect knowledge and build trusted enterprise intelligence.',
  registerSubtitle: 'Create an account and start using Enterprise AI Brain Platform.',
  noAccount: 'New to Enterprise AI Brain Platform?',
  themeLabel: 'Interface theme',
});

const themeOptions = computed<Array<{ value: ThemeMode; label: string; icon: string }>>(() => isChinese.value ? [
  { value: 'light', label: '浅色', icon: '☼' },
  { value: 'dark', label: '深色', icon: '☾' },
  { value: 'system', label: '跟随系统', icon: '▣' },
] : [
  { value: 'light', label: 'Light', icon: '☼' },
  { value: 'dark', label: 'Dark', icon: '☾' },
  { value: 'system', label: 'System', icon: '▣' },
]);

const validateLogin = () => loginFormRef.value?.validate();
const validateRegister = () => registerFormRef.value?.validate();

defineExpose({ validateLogin, validateRegister });
</script>

<style scoped>
.vone-auth-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(520px, 1.05fr) minmax(520px, 0.95fr);
  grid-template-rows: minmax(0, 1fr) 70px;
  background: var(--vone-surface-container);
  color: var(--vone-text-primary);
  overflow: hidden;
}

.vone-auth-story,
.vone-auth-workspace {
  position: relative;
  min-height: 0;
}

.vone-auth-story {
  display: flex;
  flex-direction: column;
  padding: clamp(48px, 5vw, 76px) clamp(44px, 5vw, 74px) 32px;
  background: var(--vone-auth-story-background);
  isolation: isolate;
}

.vone-auth-story__glow {
  position: absolute;
  z-index: -1;
  width: 420px;
  height: 420px;
  border-radius: 50%;
  filter: blur(14px);
  background: rgba(111, 162, 221, 0.24);
}

.vone-auth-story__glow--top { top: -210px; right: -140px; }
.vone-auth-story__glow--bottom { bottom: -240px; left: -100px; }

.vone-auth-story__copy {
  margin-top: 42px;
  color: var(--vone-text-secondary);
  font-size: clamp(17px, 1.35vw, 21px);
  line-height: 1.7;
}

.vone-auth-story__copy p { margin: 0; }

.knowledge-visual {
  position: relative;
  flex: 1;
  min-height: 380px;
  margin: 8px -28px -12px;
}

.knowledge-visual::after {
  position: absolute;
  inset: auto 2% 0;
  height: 42%;
  content: '';
  background: radial-gradient(ellipse at center, rgba(52, 121, 216, 0.18), transparent 66%);
  filter: blur(18px);
}

.knowledge-visual__links {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  overflow: visible;
}

.knowledge-visual__links path {
  fill: none;
  stroke: rgba(52, 121, 216, 0.34);
  stroke-width: 1.3;
}

.knowledge-node {
  position: absolute;
  z-index: 3;
  display: grid;
  width: 40px;
  height: 40px;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.95);
  border-radius: 50%;
  background: linear-gradient(145deg, #8bb8ef, #3479d8);
  box-shadow: 0 8px 22px rgba(32, 88, 149, 0.22), 0 0 0 5px rgba(255, 255, 255, 0.45);
  color: white;
  font-weight: 700;
}

.knowledge-node--one { left: 8%; top: 61%; }
.knowledge-node--two { left: 28%; top: 36%; }
.knowledge-node--three { left: 51%; top: 23%; }
.knowledge-node--four { right: 12%; top: 36%; }
.knowledge-node--five { right: 22%; bottom: 14%; }

.knowledge-card {
  position: absolute;
  z-index: 2;
  width: 118px;
  height: 118px;
  padding: 18px;
  border: 1px solid var(--vone-auth-glass-border);
  border-radius: 11px;
  background: var(--vone-auth-glass-background);
  box-shadow: 0 16px 38px rgba(32, 88, 149, 0.1);
  backdrop-filter: blur(8px);
  box-sizing: border-box;
}

.knowledge-card--document { left: 14%; top: 38%; }
.knowledge-card--answer { left: 46%; top: 43%; }
.knowledge-card--data { right: 8%; top: 27%; }
.knowledge-card__icon { display: block; margin-bottom: 12px; color: var(--vone-brand-500); font-weight: 800; }
.knowledge-card i { display: block; height: 5px; margin-top: 8px; border-radius: 4px; background: rgba(52, 121, 216, 0.14); }
.knowledge-card i:nth-child(3) { width: 74%; }
.knowledge-card i:nth-child(4) { width: 48%; }

.knowledge-search {
  position: absolute;
  z-index: 4;
  left: 28%;
  right: 8%;
  bottom: 23%;
  display: flex;
  align-items: center;
  gap: 14px;
  height: 58px;
  padding: 0 22px;
  border: 1px solid var(--vone-auth-glass-border);
  border-radius: 13px;
  background: var(--vone-auth-search-background);
  box-shadow: 0 14px 34px rgba(32, 88, 149, 0.13);
  color: #7890ad;
  font-size: 14px;
  backdrop-filter: blur(12px);
}

.knowledge-search__icon { color: var(--vone-brand-600); font-size: 28px; }

.vone-auth-workspace {
  display: flex;
  flex-direction: column;
  padding: 48px clamp(42px, 6vw, 82px) 28px;
  background: var(--vone-surface-container);
}

.vone-auth-toolbar {
  display: flex;
  min-height: 44px;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
}

.theme-segment,
.language-trigger {
  border: 1px solid var(--vone-border);
  border-radius: 9px;
  background: var(--vone-surface-subtle);
  box-shadow: var(--vone-shadow-sm);
}

.theme-segment {
  display: flex;
  padding: 3px;
}

.theme-segment button,
.language-trigger,
.language-dropdown button {
  border: 0;
  color: var(--vone-text-secondary);
  font: inherit;
  cursor: pointer;
}

.theme-segment button {
  display: flex;
  height: 36px;
  align-items: center;
  gap: 7px;
  padding: 0 13px;
  border-radius: 7px;
  background: transparent;
  white-space: nowrap;
}

.theme-segment button.active {
  background: var(--vone-surface-container);
  color: var(--td-brand-color);
  box-shadow: 0 2px 8px rgba(18, 61, 108, 0.1);
}

.language-switch { position: relative; }
.language-trigger { display: flex; height: 44px; align-items: center; gap: 9px; padding: 0 16px; }
.language-trigger__chevron { margin-left: 4px; }
.language-dropdown {
  position: absolute;
  z-index: 30;
  top: calc(100% + 8px);
  right: 0;
  width: 190px;
  padding: 6px;
  border: 1px solid var(--vone-border);
  border-radius: 10px;
  background: var(--vone-surface-container);
  box-shadow: var(--vone-shadow-md);
}

.language-dropdown button { display: grid; width: 100%; grid-template-columns: 24px 1fr 18px; gap: 8px; padding: 10px; border-radius: 7px; background: transparent; text-align: left; }
.language-dropdown button:hover,
.language-dropdown button.active { background: var(--td-brand-color-light); color: var(--td-brand-color); }
.language-check { text-align: right; }

.vone-auth-panel {
  display: grid;
  flex: 1;
  place-items: center;
  padding: 26px 0 10px;
}

.vone-auth-card { width: min(100%, 560px); }
.vone-auth-card__header { margin-bottom: 30px; }
.vone-auth-card__eyebrow { display: inline-block; margin-bottom: 12px; color: var(--td-brand-color); font-size: 12px; font-weight: 700; letter-spacing: 0.12em; text-transform: uppercase; }
.vone-auth-card__header h1 { margin: 0; color: var(--vone-text-primary); font-size: clamp(34px, 3vw, 46px); line-height: 1.15; letter-spacing: -0.04em; }
.vone-auth-card__header p { margin: 12px 0 0; color: var(--vone-text-secondary); font-size: 16px; line-height: 1.65; }

.vone-auth-card :deep(.t-form__item) { margin-bottom: 22px; }
.vone-auth-card :deep(.t-form__label) { padding-bottom: 8px; color: var(--vone-text-primary); font-weight: 600; }
.vone-auth-card :deep(.t-input) { min-height: 52px; border-color: var(--vone-border-strong); background: var(--vone-surface-container); }
.vone-auth-card :deep(.t-input:hover) { border-color: var(--vone-brand-400); }
.vone-auth-card :deep(.t-input--focused) { border-color: var(--td-brand-color); box-shadow: 0 0 0 3px rgba(52, 121, 216, 0.12); }
.field-icon { color: var(--vone-text-muted); font-size: 18px; }
.primary-action { height: 54px; margin-top: 4px; border-radius: 9px; font-size: 16px; font-weight: 700; }
.secondary-action { height: 52px; margin-top: 14px; border-color: var(--vone-border-strong); border-radius: 9px; color: var(--td-brand-color); font-weight: 600; }
.enterprise-icon { margin-right: 8px; }

.auth-mode-link { display: flex; align-items: center; justify-content: center; gap: 7px; margin-top: 24px; color: var(--vone-text-secondary); }
.auth-mode-link button { border: 0; background: transparent; color: var(--td-brand-color); cursor: pointer; font: inherit; font-weight: 600; }
.auth-mode-link button:disabled { cursor: not-allowed; opacity: 0.5; }

.register-grid { display: grid; grid-template-columns: 1fr 1fr; column-gap: 16px; }
.invite-banner { display: flex; gap: 12px; margin-bottom: 22px; padding: 14px 16px; border: 1px solid var(--vone-brand-200); border-radius: 10px; background: var(--vone-brand-50); color: var(--vone-text-secondary); }
.invite-banner__icon { color: var(--td-brand-color); font-size: 20px; }
.invite-banner strong { color: var(--vone-text-primary); }
.invite-banner p { margin: 3px 0 0; }
.invite-banner--error { border-color: rgba(214, 75, 85, 0.25); background: rgba(214, 75, 85, 0.08); color: var(--vone-error); }

.vone-auth-footer {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 clamp(44px, 4vw, 74px);
  border-top: 1px solid var(--vone-border);
  background: var(--vone-surface-container);
  color: var(--vone-text-secondary);
  font-size: 13px;
}

.vone-auth-footer a { color: inherit; text-decoration: none; }
.vone-auth-footer a:hover { color: var(--td-brand-color); }

@media (max-width: 1100px) {
  .vone-auth-shell { grid-template-columns: minmax(390px, 0.82fr) minmax(500px, 1.18fr); }
  .vone-auth-story { padding-inline: 42px; }
  .knowledge-card--data { display: none; }
  .knowledge-search { left: 18%; right: 4%; }
}

@media (max-width: 820px) {
  .vone-auth-shell { display: block; min-height: 100dvh; overflow: auto; background: var(--vone-surface-page); }
  .vone-auth-story { min-height: auto; padding: 28px 24px 34px; }
  .vone-auth-story :deep(.vone-brand-identity--hero) { gap: 15px; }
  .vone-auth-story :deep(.vone-brand-identity--hero .vone-brand-identity__logo-wrap) { width: 144px; }
  .vone-auth-story :deep(.vone-brand-identity--hero .vone-brand-identity__names strong) { font-size: 32px; }
  .vone-auth-story :deep(.vone-brand-identity--hero .vone-brand-identity__names small) { font-size: 18px; }
  .vone-auth-story__copy { margin-top: 20px; font-size: 15px; }
  .knowledge-visual { display: none; }
  .vone-auth-workspace { min-height: auto; padding: 22px 20px 32px; }
  .vone-auth-toolbar { order: 0; flex-wrap: wrap; }
  .theme-segment { width: 100%; }
  .theme-segment button { flex: 1; justify-content: center; padding-inline: 8px; }
  .language-switch { margin-left: auto; }
  .language-trigger { height: 38px; background: rgba(255, 255, 255, 0.72); }
  .vone-auth-panel { display: block; padding-top: 34px; }
  .vone-auth-card { width: 100%; }
  .vone-auth-card__header h1 { font-size: 34px; }
  .register-grid { grid-template-columns: 1fr; }
  .vone-auth-footer { display: flex; padding: 24px 20px; gap: 12px; border-top: 1px solid var(--vone-border); font-size: 12px; }
}

@media (max-width: 440px) {
  .vone-auth-story { padding-right: 18px; padding-left: 18px; }
  .vone-auth-workspace { padding-right: 18px; padding-left: 18px; }
  .theme-segment button { font-size: 12px; }
  .theme-segment button span { display: none; }
  .vone-auth-footer { align-items: flex-start; flex-direction: column; }
}
</style>
