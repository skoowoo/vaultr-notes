package view

// themeBootstrapScript is placed in <head> to apply the saved theme and pixel
// preference before first paint, preventing flash-of-wrong-theme.
const themeBootstrapScript = `  <script>(function(){
  var BG_DARK='#0f0f0f',BG_LIGHT='#ffffff';
  function syncViewBg(dark){if(window.vaultrDesktop&&window.vaultrDesktop.setViewBgColor)window.vaultrDesktop.setViewBgColor(dark?BG_DARK:BG_LIGHT);}
  function applyTheme(pref){var d=pref==='auto'?window.matchMedia('(prefers-color-scheme: dark)').matches:pref==='dark';document.documentElement.setAttribute('data-theme',d?'dark':'light');syncViewBg(d);}
  function applyPixel(v){document.documentElement.setAttribute('data-pixel',v==='on'?'on':'off');}
  applyTheme(localStorage.getItem('theme')||'auto');
  applyPixel(localStorage.getItem('pixel')||'off');
  window.addEventListener('storage',function(e){if(e.key==='theme')applyTheme(e.newValue||'auto');if(e.key==='pixel')applyPixel(e.newValue||'off');});
})()</script>`

// electronBootstrapScript adds the 'electron' and 'macos' classes to <html>
// when running inside the Vaultr desktop wrapper.
const electronBootstrapScript = `  <script>(function(){if(window.vaultrDesktop){document.documentElement.classList.add('electron');if(window.vaultrDesktop.platform==='darwin')document.documentElement.classList.add('macos');}})()</script>`

// electronShellSafeReloadScript defines when a full webContents reload is safe for the
// desktop multi-view shell (no editor, no filters, drawer closed, etc.) and a helper
// to refresh peer sections after vault mutations. Main reads __vaultrShellSafeForBackgroundReload via executeJavaScript.
const electronShellSafeReloadScript = `  <script>(function(){
  window.__vaultrShellSafeForBackgroundReload=function(){
    try{
      if(!window.vaultrDesktop)return true;
      var path=location.pathname||'';
      var seg=path.replace(/^\/+/,'').split('/')[0];
      if(seg==='edit')return false;
      if(seg!=='home'&&seg!=='library')return false;
      if(window.__vaultrSearchOpen)return false;
      var dr=window.__vaultrDrawer;
      if(dr&&dr.drawerOpen)return false;
      if(typeof libData!=='undefined'&&libData&&(libData.selectedTag||libData.selectedFocus))return false;
      if(seg==='home'){
        var rawTab=document.getElementById('t-raw');
        if(rawTab&&rawTab.classList.contains('on'))return false;
      }
      var st=window.__vaultrSettingsShell;
      if(st){
        if(st.saving)return false;
        if(String(st.serverUrl||'').trim()!==String(st.initialServerUrl||'').trim())return false;
      }
      return true;
    }catch(_){return false}
  };
  window.__vaultrAfterVaultMutation=async function(){
    var api=window.vaultrDesktop;
    if(api&&api.syncVaultDataAcrossSections){await api.syncVaultDataAcrossSections();return;}
    window.location.reload();
  };
})()</script>`

// themeStoreScript is the Alpine.js store body for theme and pixel management.
// Wrap it inside a document.addEventListener('alpine:init', () => { … }) call.
const themeStoreScript = `    Alpine.store('theme', {
      pref: localStorage.getItem('theme') || 'auto',
      get dark() {
        if (this.pref === 'auto') return window.matchMedia('(prefers-color-scheme: dark)').matches;
        return this.pref === 'dark';
      },
      set(pref) {
        this.pref = pref;
        localStorage.setItem('theme', pref);
        this._apply();
      },
      _apply() {
        var dark = this.dark;
        document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'light');
        if (window.vaultrDesktop && window.vaultrDesktop.setViewBgColor)
          window.vaultrDesktop.setViewBgColor(dark ? '#0f0f0f' : '#ffffff');
      },
      init() {
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
          if (this.pref === 'auto') this._apply();
        });
      }
    });
    Alpine.store('pixel', {
      enabled: localStorage.getItem('pixel') === 'on',
      toggle() {
        this.enabled = !this.enabled;
        const v = this.enabled ? 'on' : 'off';
        localStorage.setItem('pixel', v);
        document.documentElement.setAttribute('data-pixel', v);
      },
      init() {
        document.documentElement.setAttribute('data-pixel', this.enabled ? 'on' : 'off');
      }
    });
    Alpine.store('settingsModal', { open: false });`
