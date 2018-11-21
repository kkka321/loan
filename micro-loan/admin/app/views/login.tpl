<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Rupiah Cepat</title>
    <style>
        body,html{
            width: 100%;
            height: 100%;
            background-color: rgb(66,133,244);
        }
        .clear{overflow: hidden;}
        .fl{float: left;}
        .container{
            width: 1190px;
            margin: 0 auto;
            position: absolute;
            top: 48%;
            left: 50%;
            transform: translate(-50%, -50%);
        }
        .bgLogo{
           width: 558px;height: 455px;
        }
        .bgLogo img{width: 100%;}
        .form{
            width: 618px;
            height: 455px;
            background-color: #fff;
            box-shadow: 10px 10px 16px 0 rgba(49,99,181,0.34);
            border-radius: 4px;
        }

        .wel{
            font-size: 30px;
            color: #4285F4;
            letter-spacing: 0;
            text-align: center;
            font-weight: bold;
        }
        form{
            padding: 0 54px;
        }
        .inputDiv{
            position: relative;
            width: 100%;
            height: 50px;
            margin: 0 auto;
            border-radius: 4px;
            border: 1px solid #4285F4;
            margin-bottom: 20px;
        }
        input{
            width: 95.5%;
            height: 95.5%;
            padding-left: 22px;
            line-height: 100%;
            position: absolute;
            left: 0;
            top: 0;
            z-index: 10;
            color: #4285F4;
            background-color: transparent;
            border: none;
            outline: none;
            border-radius: 4px;
        }
        input:-webkit-autofill { 
            background-color: transparent !important;
            box-shadow: 0 0 0px 1000px transparent inset !important;
        }
        input::-webkit-input-placeholder {
            font-size: 14px;
            color: #4A4A4A;
            opacity: 0.4;
        }
        .inputDiv.code{
            width: 300px;
            float: left;
        }
        .code input{
            width: 92.7%;
        }
        .codeImg{
            width: 184px;
            height: 52px;
            margin-left: 24px;
            display: inline-block;
            background: #D8D8D8;
            border-radius: 4px;
            cursor: pointer;
        }
        h4{
            opacity: 0.4;
            font-family: Roboto-Light;
            font-size: 14px;
            color: #9B9B9B;
            letter-spacing: 0;
            font-weight: normal;
            text-align: center;
            clear: both;
        }
        .captcha-img{
            cursor: pointer;
            width: 100%;
        }
        .box-footer{
            margin-top: 8%;
        }
        .btn{
            width:  100%;
            height: 50px;
            background-color: #4285F4;;
            margin: 0 auto;
            line-height: 36px;
            text-align: center;
            color:   #FFF;
            border: none;
            outline: none;
            display: block;
            border-radius: 4px;
            cursor: pointer;
            font-size: 24px;
            color: #fff;
            letter-spacing: 0;
        }
    </style>
</head>
<body>
    <div class="container clear">
        <div class="bgLogo fl">
            <img src="https://download.rupiahcepatweb.com/static/img/bg-logo.png" alt="">
        </div>
        <div class="form fl">
            <p class="wel">Welcome to KUFI !</p>
            <form autocomplete="off" method="post" action="/login_confirm">
                <div class="box-body">
                  <div class="form-group inputDiv">
                    <input name="email" type="email" class="form-control" id="email" required="" placeholder="Input Your Login ID" autocomplete="off">
                  </div>
                  <div class="form-group inputDiv">
                    <input name="password" type="password" class="form-control" id="password" required="" placeholder="Input Your Password">
                  </div>
                </div>
                 <div class="form-group inputDiv code">
                    <input name="captcha" type="text" required="" autocomplete="off" placeholder="Security Code">
                 </div>
                 <div class="codeImg">
                 {{create_captcha}}
                 </div>
               <h4>Can't read the image?Click it get a new one</h4>
                <div class="box-footer">
                  <button type="submit" class="btn btn-primary">Login</button>
                </div>
              </form>
        </div>
    </div>
</body>
</html>
