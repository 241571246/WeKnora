(function () {
  const root = document.documentElement;
  const body = document.body;
  const toast = document.getElementById("toast");
  const themeMedia = window.matchMedia("(prefers-color-scheme: dark)");
  let toastTimer;

  function showToast(message) {
    window.clearTimeout(toastTimer);
    toast.textContent = message;
    toast.classList.add("show");
    toastTimer = window.setTimeout(() => toast.classList.remove("show"), 2600);
  }

  function normalizeRoute() {
    const candidate = window.location.hash.replace("#", "");
    return ["login", "register", "home"].includes(candidate) ? candidate : "login";
  }

  function applyRoute() {
    const route = normalizeRoute();
    body.dataset.route = route;
    const lang = root.lang === "en" ? "en" : "zh";
    const titles = {
      login: lang === "zh" ? "登录 · VONE知识库" : "Sign in · Vone Knowledge",
      register: lang === "zh" ? "注册 · VONE知识库" : "Register · Vone Knowledge",
      home: lang === "zh" ? "知识库 · VONE知识库" : "Knowledge bases · Vone Knowledge"
    };
    document.title = titles[route];
    body.querySelector(".home-shell")?.classList.remove("sidebar-open");
  }

  function resolvedTheme(value) {
    return value === "system" ? (themeMedia.matches ? "dark" : "light") : value;
  }

  function applyTheme(value) {
    const selected = ["light", "dark", "system"].includes(value) ? value : "light";
    root.dataset.theme = resolvedTheme(selected);
    root.dataset.themePreference = selected;
    document.querySelectorAll("[data-theme-value]").forEach((button) => {
      button.setAttribute("aria-pressed", String(button.dataset.themeValue === selected));
    });
    document.getElementById("homeThemeButton").textContent = selected === "dark" ? "◐" : selected === "system" ? "▣" : "☀";
    try { localStorage.setItem("vone-prototype-theme", selected); } catch (_) {}
  }

  function applyLanguage(lang) {
    const selected = lang === "en" ? "en" : "zh";
    const copyKey = selected === "en" ? "en" : "cn";
    root.lang = selected === "en" ? "en" : "zh-CN";
    document.querySelectorAll("[data-cn][data-en]").forEach((node) => {
      node.textContent = node.dataset[copyKey];
    });
    document.querySelectorAll("[data-placeholder-cn][data-placeholder-en]").forEach((node) => {
      node.placeholder = node.dataset[`placeholder${selected === "en" ? "En" : "Cn"}`];
    });
    document.querySelectorAll("[data-lang-value]").forEach((button) => {
      button.setAttribute("aria-pressed", String(button.dataset.langValue === selected));
    });
    try { localStorage.setItem("vone-prototype-lang", selected); } catch (_) {}
    applyRoute();
  }

  window.addEventListener("hashchange", applyRoute);
  themeMedia.addEventListener?.("change", () => {
    if (root.dataset.themePreference === "system") applyTheme("system");
  });

  document.addEventListener("click", (event) => {
    const themeButton = event.target.closest("[data-theme-value]");
    if (themeButton) {
      applyTheme(themeButton.dataset.themeValue);
      document.getElementById("themePopover").classList.remove("open");
      return;
    }
    const langButton = event.target.closest("[data-lang-value]");
    if (langButton) {
      applyLanguage(langButton.dataset.langValue);
      return;
    }
    const passwordButton = event.target.closest(".password-toggle");
    if (passwordButton) {
      const input = passwordButton.parentElement.querySelector("input");
      input.type = input.type === "password" ? "text" : "password";
      passwordButton.textContent = input.type === "password" ? "◎" : "◉";
      return;
    }
    const demo = event.target.closest("[data-demo-message]");
    if (demo) {
      if (demo.classList.contains("card-menu")) event.stopPropagation();
      showToast(demo.dataset.demoMessage);
    }
  });

  document.getElementById("loginForm").addEventListener("submit", (event) => {
    event.preventDefault();
    window.location.hash = "home";
  });
  document.getElementById("registerForm").addEventListener("submit", (event) => {
    event.preventDefault();
    showToast(root.lang === "en" ? "Prototype account created. Opening workspace…" : "原型账户创建成功，正在进入工作区…");
    window.setTimeout(() => { window.location.hash = "home"; }, 500);
  });

  const dialog = document.getElementById("newKbDialog");
  document.getElementById("newKbButton").addEventListener("click", () => dialog.showModal());
  document.getElementById("newKbForm").addEventListener("submit", (event) => {
    const submitter = event.submitter;
    if (submitter?.value === "create") showToast(root.lang === "en" ? "Prototype created successfully." : "知识库原型创建成功。");
  });
  document.getElementById("homeThemeButton").addEventListener("click", (event) => {
    event.stopPropagation();
    document.getElementById("themePopover").classList.toggle("open");
  });
  document.addEventListener("click", () => document.getElementById("themePopover").classList.remove("open"));
  document.getElementById("mobileMenuButton").addEventListener("click", () => body.querySelector(".home-shell").classList.toggle("sidebar-open"));

  let storedTheme = "light";
  let storedLang = "zh";
  try {
    storedTheme = localStorage.getItem("vone-prototype-theme") || "light";
    storedLang = localStorage.getItem("vone-prototype-lang") || "zh";
  } catch (_) {}
  applyTheme(storedTheme);
  applyLanguage(storedLang);
  applyRoute();
})();
