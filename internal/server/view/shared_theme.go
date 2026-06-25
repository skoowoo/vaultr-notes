package view

// themeBootstrapScript is placed in <head> to apply the neo theme before first
// paint. Hardcoded to neo — the only supported theme.
const themeBootstrapScript = `  <script>(function(){
  try{if(localStorage.getItem('theme')==='pixel')localStorage.setItem('theme','neo');}catch(_){}
  document.documentElement.setAttribute('data-theme','neo');
  if(window.vaultrDesktop&&window.vaultrDesktop.setViewBgColor)window.vaultrDesktop.setViewBgColor('#ffffff');
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

// alpineStoresScript initializes all Alpine.js global stores.
// Wrap it inside a document.addEventListener('alpine:init', () => { … }) call.
const alpineStoresScript = `    Alpine.store('settingsModal', { open: false });`
