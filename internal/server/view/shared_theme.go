package view

// themeBootstrapScript is placed in <head>, before first paint. Dark-mode
// tokens live on bare :root (shared_tokens.go) — Linear's near-black canvas
// is the shipped default — with the light adaptation applying via
// :root[data-theme="light"] or an unforced OS light preference. There's no
// user-facing toggle yet (TODO: once one exists, read the stored preference
// here and set data-theme explicitly before first paint instead of relying
// on matchMedia), so this only has to agree with the same
// prefers-color-scheme check the CSS itself uses.
const themeBootstrapScript = `  <script>(function(){
  if(window.vaultrDesktop&&window.vaultrDesktop.setViewBgColor){
    var light=window.matchMedia&&window.matchMedia('(prefers-color-scheme: light)').matches;
    window.vaultrDesktop.setViewBgColor(light?'#fcfcfc':'#010102');
  }
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
      if(seg!=='home')return false;
      if(window.__vaultrSearchOpen)return false;
      var dr=window.__vaultrDrawer;
      if(dr&&dr.drawerOpen)return false;
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
