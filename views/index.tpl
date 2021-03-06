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
    <p>输入0,准备.输入1,A选择完毕.输入2,B选择错误.输入3,B选择正确,20秒后游戏结束收到gameover事件</p>
    <p>join room address:{{.webHost}}/?chat={{.Room}}</p>
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
      var socket = io("{{.gameHost}}/?chat={{.Room}}&nick=MaxNick&head=MaxHead");
      var counter = 4 //题号

      $('form').submit(function(){
        input = $('#m').val()
        if(input =="0"){
            console.log("client confirmed")
            socket.emit('confirm');
        }


        if(input == "1"){
            console.log("client:第"+counter+"题,A选择完毕，等待 B")
            console.log("广播到全局参数:"+counter+";")
            socket.emit('asend',counter+";");
        }
        if(input =="2"){
            console.log("client:B 选择了第"+counter+"题的错误答案的发送bsend事件,等待A选择")
            socket.emit('bsend',counter+",0");
            counter++
        }
        if(input =="3"){
            console.log("client:B 选择了第"+counter+"题的正确答案的,发送bsend事件,等待A选择")
            socket.emit('bsend',counter+",1");
            counter++
        }

        if(input =="gameover"){
            counter = 0
            socket.emit('gameover')
        }

        //console.log(msg)
        //socket.emit('chatmsg', msg);

        $('#m').val('');
        return false;
      });


      socket.on('pendb', function(msg){
        console.log("收到服务器pendb事件参数"+msg+"(A选择了第"+msg+"题)")
        console.log("等待B选择")
      });

      socket.on('penda', function(msg){
        console.log("收到服务器penda事件参数选择结果:"+msg)
        console.log("等待a选择")
      });

      socket.on('user', function(msg){
        console.log("接收到对方玩家信息:"+msg)
      });

      socket.on('gameover', function(msg){
        counter = 0
        console.log("client:recived gameover Event from Server:"+msg)
        console.log("game over")
      });
      socket.on('chatmsg', function(msg){
        console.log(msg)
      });

      socket.on('ready', function(msg){
        console.log(msg)
      });

      socket.on('gamestart', function(msg){
        console.log("clent:recived Event from Server:Game start")
      });


      socket.on('joined', function(msg){
        console.log(msg)
        $('#qrcodeCanvas').remove()
      });

      socket.on('info', function(msg){
        console.log(msg)
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
