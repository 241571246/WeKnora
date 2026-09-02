import faviconUrl from './assets/favicon.ico';
import logoBlueUrl from './assets/vone-logo-blue.png';
import logoLightUrl from './assets/vone-logo-light.png';

export type BrandProfile = 'upstream' | 'vone';

export const brandProfile: BrandProfile =
  import.meta.env.VITE_BRAND_PROFILE === 'upstream' ? 'upstream' : 'vone';

export const voneBrand = {
  id: 'vone',
  version: '0.1.0',
  productNameZh: 'AI大脑平台',
  productNameEn: 'AI Brain Platform',
  shortName: 'VONE',
  copyright: '@vonechina 2026',
  upstreamAttribution: 'Powered by Tencent WeKnora',
  upstreamUrl: 'https://github.com/Tencent/WeKnora',
  description: '连接知识、检索答案，构建可信的企业智能。',
  logos: {
    blue: logoBlueUrl,
    light: logoLightUrl,
  },
  favicon: faviconUrl,
} as const;

const ensureLink = (rel: string, href: string) => {
  let link = document.head.querySelector<HTMLLinkElement>(`link[rel="${rel}"]`);
  if (!link) {
    link = document.createElement('link');
    link.rel = rel;
    document.head.appendChild(link);
  }
  link.href = href;
};

let themeObserver: MutationObserver | null = null;

const syncBrandCanvas = () => {
  const isDark = document.documentElement.getAttribute('theme-mode') === 'dark';
  const color = isDark ? '#0b1320' : '#f4f7fb';
  const rgb = isDark ? [11, 19, 32] : [244, 247, 251];

  document.documentElement.style.background = color;
  if (document.body) document.body.style.background = color;
  const app = document.getElementById('app');
  if (app) app.style.background = color;

  const runtime = (window as unknown as {
    runtime?: { WindowSetBackgroundColour?: (r: number, g: number, b: number, a: number) => void };
  }).runtime;
  runtime?.WindowSetBackgroundColour?.(rgb[0], rgb[1], rgb[2], 255);
};

export const applyVoneBrand = () => {
  if (brandProfile !== 'vone') return;

  const root = document.documentElement;
  root.dataset.brandProfile = voneBrand.id;
  root.dataset.brandVersion = voneBrand.version;
  document.title = `${voneBrand.productNameZh} · ${voneBrand.productNameEn}`;

  const description = document.head.querySelector<HTMLMetaElement>('meta[name="description"]');
  if (description) description.content = voneBrand.description;

  ensureLink('icon', voneBrand.favicon);
  ensureLink('shortcut icon', voneBrand.favicon);
  ensureLink('apple-touch-icon', voneBrand.favicon);
  syncBrandCanvas();

  themeObserver?.disconnect();
  themeObserver = new MutationObserver(syncBrandCanvas);
  themeObserver.observe(root, { attributes: true, attributeFilter: ['theme-mode'] });
};
