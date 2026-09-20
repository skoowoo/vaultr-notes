package view

// themeBootstrapScript runs in <head> before first paint, so there's no
// flash of the wrong theme. Dark lives on bare :root (shared_tokens.go);
// light applies via :root[data-theme="light"] or an unforced OS preference.
//
// The Settings → Editor "Theme" toggle stores light/dark/auto under the
// 'vaultr-theme' localStorage key and calls window.__vaultrApplyTheme (see
// shared_settings_modal.go's setTheme()); "auto" also gets a live
// prefers-color-scheme listener so it follows an OS change without a reload.
const themeBootstrapScript = `  <script>(function(){
  window.__vaultrApplyTheme=function(pref){
    try{
      var sysLight=window.matchMedia&&window.matchMedia('(prefers-color-scheme: light)').matches;
      var effectiveLight=pref==='light'?true:(pref==='dark'?false:sysLight);
      if(pref==='light'){document.documentElement.setAttribute('data-theme','light');}
      else if(pref==='dark'){document.documentElement.setAttribute('data-theme','dark');}
      else{document.documentElement.removeAttribute('data-theme');}
      if(window.vaultrDesktop&&window.vaultrDesktop.setViewBgColor){
        window.vaultrDesktop.setViewBgColor(effectiveLight?'#f9f9fb':'#18191e',pref==='light'?'light':(pref==='dark'?'dark':''));
      }
      if(window.__vaultrApplyAccent)window.__vaultrApplyAccent();
    }catch(_){}
  };
  var stored=localStorage.getItem('vaultr-theme');
  window.__vaultrApplyTheme(stored==='light'||stored==='dark'?stored:'auto');
  if(window.matchMedia){
    window.matchMedia('(prefers-color-scheme: light)').addEventListener('change',function(){
      var cur=localStorage.getItem('vaultr-theme');
      if(cur!=='light'&&cur!=='dark')window.__vaultrApplyTheme('auto');
    });
  }
})()</script>`

// electronBootstrapScript adds the 'electron' and 'macos' classes to <html>
// when running inside the Vaultr desktop wrapper.
const electronBootstrapScript = `  <script>(function(){if(window.vaultrDesktop){document.documentElement.classList.add('electron');if(window.vaultrDesktop.platform==='darwin')document.documentElement.classList.add('macos');}})()</script>`

// electronShellSafeReloadScript defines when a full webContents reload is
// safe for the desktop multi-view shell, and a helper to refresh peer
// sections after vault mutations. Main reads
// __vaultrShellSafeForBackgroundReload via executeJavaScript.
const electronShellSafeReloadScript = `  <script>(function(){
  window.__vaultrShellSafeForBackgroundReload=function(){
    try{
      if(!window.vaultrDesktop)return true;
      var path=location.pathname||'';
      var seg=path.replace(/^\/+/,'').split('/')[0];
      if(seg==='edit')return false;
      if(seg!=='home')return false;
      if(window.__vaultrSearchOpen)return false;
      var cp=window.__vaultrContentPane;
      if(cp&&cp.contentPaneOpen)return false;
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

// alpineStoresScript initializes all Alpine.js global stores.
// Wrap it inside a document.addEventListener('alpine:init', () => { … }) call.
const alpineStoresScript = `    Alpine.store('settingsModal', { open: false });`
