
  function foldersCtrl() {
    return Object.assign(drawerCtrl(), {
      init() {
        this.initDrawer();
      },
    });
  }
