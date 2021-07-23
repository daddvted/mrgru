jQuery(function ($) {
  var websshLoginFormID = '#webssh-login-form',
    websshLoginBtn = $('#webssh-login-btn'),
    info = $('#info'),
    toolbar = $('#toolbar'),
    menu = $('#menu'),
    progress = $("#progress"),
    cell = {},
    titleElement = document.querySelector('title'),
    customizedFont = "Hack",  // Named by style.css
    fields = ["hostname", "port", "username", "password"],
    defaultTitle = "Term1nal",
    currentTitle = undefined,
    term = new Terminal();


  // Hide toolbar first
  toolbar.hide();
  menu.hide();
  // popupForm.hide();

  function setMsg(text) {
    $('#msg').html(text);
  }

  function copySelectedText() {
    let el = document.createElement('textarea');
    el.value = term.getSelection();
    el.select();
    document.execCommand('copy');
  }


  // Maybe cancel this after using direct upload/download
  function setSession(name, data) {
    window.sessionStorage.clear()
    window.sessionStorage.setItem(name, data)
  }

  function getSession(name) {
    return window.sessionStorage.getItem(name)
  }

  function validateFormData(formID, formFields) {
    let data = new FormData(document.querySelector(formID));
    let result = { error: "" }

    formFields.forEach(function (attr) {
      var val = data.get(attr)
      if (!val) {
        result.error = `${attr} is required`;
        return result;
      }
      // } else {
      //   result[attr] = val;
      // }
    })

    // Set current tab title
    currentTitle = `${data.get('username')}@port:${data.get('port')}`;
    return result;
  }

  function getCurrentDimension(term) {
    if (!cell.width || !cell.height) {
      try {
        cell.width = term._core._renderService._renderer.dimensions.actualCellWidth;
        cell.height = term._core._renderService._renderer.dimensions.actualCellHeight;
      } catch (error) {
        console.log("Error getting curent Dimension")
      }
    }

    let cols = parseInt(window.innerWidth / cell.width, 10),
      rows = parseInt(window.innerHeight / cell.height, 10);
    return [cols, rows];
  }

  function resizeTerminal(term) {
    let dim = getCurrentDimension(term);
    term.resizeWindow(dim[0], dim[1]);
  }

  // Use window.Textdecoder to process terminal data from server,
  // then write to Xterm.js
  function processBlobData(blob, callback, decoder) {
    if (window.TextDecoder) {
      let reader = new window.FileReader();

      reader.onload = function () {
        let text;
        try {
          text = decoder.decode(reader.result);
        } catch (err) {
          console.log(`!!! Decode error: ${err}`);
        } finally {
          callback(text);
        }

      }
      reader.onerror = function (err) {
        console.log(`Filereader onerror: ${err}`)
      }
      reader.readAsArrayBuffer(blob);
    } else {
      console.log("!!! Browser does not support TextDecoder");
    }
  }


  function connectCallback(resp) {

    if (resp.status !== 200) {
      setMsg(`${resp.status}: ${resp.statusText}`);
      return;
    }

    let defaultEncoding = 'utf-8',
      msg = resp.responseJSON;

    if (!msg.id) {
      err = msg.status.toLowerCase();
      setMsg(msg.status);
      return;
    } else {
      setSession("minion", msg.id)
    }

    if (!msg.encoding) {
      // Use default encoding when unable to detect serer encoding
      // msg.encoding = defaultEncoding;
      console.log(`Use default encoding: ${defaultEncoding}`);
      var decoder = defaultEncoding;
    } else {
      console.log(`Server encoding : ${msg.encoding}`);
      try {
        var decoder = new window.TextDecoder(msg.encoding);
      } catch (EncodingError) {
        console.log(`Unknown encoding: ${msg.encoding}`);
      }
    }

    // Prepare websocket
    let proto = window.location.protocol,
      url = window.location.href,
      char = (proto === "http:" ? "ws:" : "wss:"),
      wsURL = `${url.replace(proto, char)}ws?id=${msg.id}`,
      sock = new window.WebSocket(wsURL),
      terminal = document.getElementById("terminal"),
      term = new window.Terminal({
        cursorBlink: true,
        theme: {
          background: "black"
        }
      });
    term.fitAddon = new window.FitAddon.FitAddon();
    term.loadAddon(term.fitAddon);


    function write2terminal(text) {
      if (term) {
        term.write(text);
        if (!term.resized) {
          resizeTerminal(term);
          term.resized = true;
        }
      }
    }

    term.resizeWindow = function (cols, rows) {
      if (cols !== this.cols || rows !== this.rows) {
        console.log('Resizing terminal to geometry: ' + JSON.stringify({ 'cols': cols, 'rows': rows }));
        this.resize(cols, rows);
        sock.send(JSON.stringify({ 'resize': [cols, rows] }));
      }
    };

    term.onData(function (data) {
      sock.send(JSON.stringify({ 'data': data }));
    });

    // Copy on selection
    window.addEventListener('mouseup', copySelectedText);

    sock.onopen = function () {
      menu.show();

      term.open(terminal);

      //Full screen
      $('#terminal .terminal').toggleClass('fullscreen');
      term.fitAddon.fit();

      term.setOption('fontFamily', customizedFont);
      term.focus();
      // titleElement.text = tmpData.title || defaultTitle;
      titleElement.text = currentTitle || defaultTitle;
    };

    sock.onmessage = function (msg) {
      processBlobData(msg.data, write2terminal, decoder);
    };

    sock.onerror = function (event) {
      console.error(event);
    };

    sock.onclose = function (event) {
      // Hide toolbar again
      toolbar.hide();
      menu.hide();

      sock = undefined;
      term.dispose();
      term = undefined;
      setMsg(event.reason);
      titleElement.text = defaultTitle;

      // Remove some event listeners
      window.removeEventListener("mouseup", copySelectedText);
    };

    $(window).resize(function () {
      if (term) {
        resizeTerminal(term);
      }
    });
  } // ajaxCallback()

  function connect(formID) {

    let data = new FormData(document.querySelector(formID))

    $.ajax({
      url: '/',
      type: 'post',
      data: JSON.stringify(Object.fromEntries(data)),
      // data: data,
      complete: connectCallback,
      error: function (err) {
        console.log("wtf");
        console.log(err)
      },
      cache: false,
      contentType: false,
      processData: false
    });
  } // connect function


  // ====================================
  // Setup Event Listeners
  // ====================================
  websshLoginBtn.click(function (event) {
    event.preventDefault();
    // websshLoginFormID[0].checkValidity();
    // Clean msg
    setMsg("");
    let result = validateFormData(websshLoginFormID, fields);
    if (result.error) {
      setMsg(result.error);
    } else {
      connect(websshLoginFormID);
      console.log(result);
    }
  });

  $("#upload").click(function () {
    // Clean this for triggering change event for same file
    this.value = "";
    // Clean info text
    info.text("");
  });

  // Listen to "file" change event to upload file,
  // monitor "progress" event to calculate uploading percentage
  $("#upload").change(function () {
    console.log(this);
    var file = this.files[0]
    console.log(file);
    var formData = new FormData()
    // formData.append("minion", getSession("minion"))
    formData.append("upload", file)


    $.ajax({
      url: `/upload?minion=${getSession("minion")}`,
      type: "POST",
      data: formData,
      cache: false,
      contentType: false,
      processData: false,
      timeout: 600000,
      async: true,

      xhr: function () {
        var theXHR = $.ajaxSettings.xhr();
        if (theXHR.upload) {
          progress.show();
          theXHR.upload.addEventListener('progress', function (e) {
            if (e.lengthComputable) {
              percent = Math.ceil(e.loaded / e.total * 100);
              // console.log(percent);
              $(progress).attr("value", percent);
              if (percent == 100) {
                progress.hide();
              }
            }
          }, false);
        }
        return theXHR;
      },
      success: function (data) {
        info.text(data);
      },
      error: function (error) {
        progress.hide()
        console.log(error)
      }
    }); //.ajax()
  }); // #upload.change()

  $("#download").click(function () {
    file = $("#downloadFile").val()
    if (file === "") {
      alert("Input file path")
      return
    }

    // Chrome save dialog will open after file has been downloaded
    // fetch(`download?filepath=${file}&minion=${getSession("minion")}`)
    // .then((resp) =>{
    //   if (!resp.ok) {
    //     alert(`${file} not exist`)
    //   } else {
    //     resp.blob().then((blob) => {
    //       let url = window.URL.createObjectURL(blob);
    //       let a = document.createElement('a');
    //       a.style.display = 'none';
    //       a.href = url;
    //       a.download = file.split('/').pop();
    //       document.body.appendChild(a);
    //       a.click();
    //       window.URL.revokeObjectURL(url)
    //     })
    //   }
    // })
    // .catch((err) => {
    //   alert(err)
    // })

    // With Chrome download progress
    // window.location.href = `download?filepath=${file}&minion=${getSession("minion")}`;
    window.open(`download?filepath=${file}&minion=${getSession("minion")}`);
  }); // #download.click()

  menu.click(function () {
    $("#downloadFile").val("");
    progress.hide();
    toolbar.toggle();
    info.text("")
  })

  $(window).on('beforeunload', function (evt) {
    console.log(evt);
    // Use 'beforeunload' to prevent "ctrl+W" from closing browser tab
    return "bye";
  });

}); // jQuery
