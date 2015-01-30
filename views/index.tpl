<!doctype html>
<html>
  <head>
    <title>Socket.IO chat</title>
    <meta name="viewport" content="width=device-width,maximum-scale=1,minimum-scale=1">
    <style>
      * { margin: 0; padding: 0; box-sizing: border-box; }
      body { font: 13px Helvetica, Arial; }
      form { background: #000; padding: 3px; position: fixed; bottom: 0; width: 100%; }
      form input { border: 0; padding: 10px; width: 80%; margin-right: .5%; }
      form button { width: 9%; background: rgb(130, 224, 255); border: none; padding: 10px; }
      #messages { list-style-type: none; margin: 0; padding: 0; }
      #messages li { padding: 5px 10px; }
      #messages li:nth-child(odd) { background: #eee; }
    </style>
  </head>
  <body>
    <ul id="messages"></ul>
    <form action="">
      <input id="m" autocomplete="off" /><button>Send</button>
    </form>
    <div id="qrcodeCanvas"></div>
    <script src="/static/js/socket.io.js"></script>
    <script src="/static/js/jquery-1.11.1.js"></script>
    <script src="/static/js/jquery-1.11.1.js"></script>
    <script src="/static/js/jquery.cookie.js"></script>
    <script>
    function getUrlParam(name)
    {
        var reg = new RegExp("(^|&)"+ name +"=([^&]*)(&|$)"); //构造一个含有目标参数的正则表达式对象
            var r = window.location.search.substr(1).match(reg);  //匹配目标参数
                if (r!=null) return unescape(r[2]); return null; //返回参数值
    }
    </script>
    <script>
      var socket = io("{{.gameHost}}/?chat={{.Room}}");

      $('form').submit(function(){
        socket.emit('chat message', $('#m').val());
        //$('#messages').append($('<li>').text($('#m').val()));
        $('#m').val('');
        return false;
      });

      socket.on('chat message', function(msg){
        $('#messages').append($('<li>').text(msg));
      });

      socket.on('connect', function(msg){
      });

      socket.on('joined', function(msg){
        $('#messages').append($('<li>').text(msg));
        $('#qrcodeCanvas').remove()
      });

      socket.on('info', function(msg){
        $('#messages').append($('<li>').text(msg));
      });

        if(getUrlParam('chat')){
            socket.on('connect', function(msg){
            $('#qrcodeCanvas').remove()
            });
        }


    </script>
    <script type="text/javascript" src="/static/js/jquery.qrcode.js"></script>
    <script type="text/javascript" src="/static/js/qrcode.js"></script>
        <script>
	        jQuery('#qrcodeCanvas').qrcode({
		    text	: "{{.webHost}}/?chat={{.Room}}"
	    });
    </script>
  </body>
</html>
