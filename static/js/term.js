// Main
jQuery(function ($) {

  var terminal = new Terminal({
    screenKeys: true,
    useStyle: true,
    cursorBlink: true,
    fullscreenWin: true,
    maximizeWin: true,
    screenReaderMode: true,
    cols: 128,
  });

  var protocol = (location.protocol === "https:") ? "wss://" : "ws://";
  var url = protocol + location.host + "/ws";
  var ws = new WebSocket(url);

  var attachAddon = new AttachAddon.AttachAddon(ws);
  var fitAddon = new FitAddon.FitAddon();
  // var webLinksAddon = new WebLinksAddon.WebLinksAddon();
  // var unicode11Addon = new Unicode11Addon.Unicode11Addon();
  // var serializeAddon = new SerializeAddon.SerializeAddon();
  terminal.loadAddon(fitAddon);
  // terminal.loadAddon(webLinksAddon);
  // terminal.loadAddon(unicode11Addon);
  // terminal.loadAddon(serializeAddon);

  ws.onopen = function () {
    terminal.open(document.getElementById("terminal"));
    terminal.loadAddon(attachAddon);
    terminal._initialized = true;
    // fitAddon.fit();
    $("#terminal .terminal").toggleClass("fullscreen");
    terminal.focus();

    setTimeout(function () { fitAddon.fit() });
    terminal.onResize(function (event) {
      var rows = event.rows;
      var cols = event.cols;
      var size = JSON.stringify({ cols: cols, rows: rows + 1 });
      var send = new TextEncoder().encode("\x01" + size);
      console.log("resize to ", size);
      ws.send(send);
    });
    terminal.onTitleChange(function (event) {
      console.log(event)
    });
    window.onresize = function () {
      fitAddon.fit();
    };
  };


}) //jQuery
