

/*
公共方法ajax_post方法,ajax 提交默认为是post 提交
参数说明
 1.url  是需要提交的url地址
 2.需要提交的数据，建议以formdata格式的，数据valaue 要转json
    var formData=new FormData()
    formData.append("password", JSON.stringify(password));
 3.回掉函数，调用成功后需要执行的函数。
 4.是否添加确认提示，添加 为 true,不添加为false
 5.添加的提示信息

 */
function common_ajax_post (url,data,success_callback,confirm,confirm_infotext) {
    if (confirm&&confirm_infotext) {
        common_sbumit_confirm(confirm_infotext,
            function (is_confirm) {
                if (!is_confirm) {
                    return false;
                }

                $.ajax({
                    type: "post",
                    url: url,
                    data: data,
                    processData:false,
                    contentType:false,
                    success: function (result) {
                        if (result.code ==200) {
                            common_suucess(result.message)
                            success_callback()

                        } else {
                            common_error(result.message)
                        }
                    },


                });
            });

    };

}
/*
获取焦点公共事件
1.元素id 或者class
2.获取焦点后需要执行的函数

 */
function Get_focus(IdOrClass,call_back_function) {

    $('body').on('mouseover',IdOrClass,call_back_function)

};
//失去焦点公共事件
function Get_focus(IdOrClass,call_back_function) {

    $('body').on('mouseout',IdOrClass,call_back_function)

};
/**
 * 获取当前宽度和屏幕总宽度，弹层居中
 * @param info
 * 1.传入jqeury封装过的元素
 *   例子：scree_center($("#id))
 */
function scree_center(popupName) {
    var _scrollHeight = $(document).scrollTop(),
        //获取当前窗口距离页面顶部高度
        _windowHeight = $(window).height(),
        //获取当前窗口高度
        _windowWidth = $(window).width(),
        //获取当前窗口宽度
        _popupHeight = popupName.height(),
        //获取弹出层高度
        _popupWeight = popupName.width(); //获取弹出层宽度
    _posiTop = (_windowHeight - _popupHeight) / 2 + _scrollHeight;
    _posiLeft = (_windowWidth - _popupWeight) / 2;
    popupName.css({
        "left": _posiLeft + "px",
        "display": "block"
    }); //设置position
}



$(document).ready(function() {
    // 刷新当前页
    $('.reload-page').click(function(){
        location.reload();
    });

    $('#select_lang').change(function() {
        var lang = $(this).val();
        var url = window.location.href;

        if (-1 != url.indexOf('lang')) {
            url = url.replace(/(&?)lang=(\w+)-(\w+)/g, '');
        }

        if (-1 != url.indexOf('?')) {
            if ('?' == url.substr(url.length - 1, 1)) {
                url = url + 'lang=' + lang;

            } else {
                url = url + '&lang=' + lang;
            }
        } else {
            url =  url + '?lang=' + lang;
        }

        window.location = url;
    });
});
$(function(){
    var url = location.href
    var newurl=""
    if (-1 != url.indexOf('?')) {
        if ('&' == url.substr(url.length - 1, 1)) {
            newurl = url

        } else {
            newurl = url
        }
    }
    if (newurl.indexOf("/customer/detail")!=-1||newurl.indexOf("/riskctl/phone_verify")!=-1){
        if($(".pull-right").children("li").length>5){
            if($(".pull-right").children("li").length>5){
                url= $(".pull-right li:nth-child(5) a").attr("href")
                url=url+"?p="
                $(".pull-right li:nth-child(5) a").attr("href",url)
                $(".pull-right li:nth-child(3) a").attr("href",url)
            }
        }
        if (newurl.indexOf("&sort")!=-1 ||newurl.indexOf("&p")!=-1||newurl.indexOf("&p")!=-1){
            $("#tab5").addClass("active")
            $("#tab_5").addClass("active")
        }else {
            $("#tab1").addClass("active")
            $("#tab_1").addClass("active")
        }
        if (newurl.indexOf("/riskctl/phone_verify") != -1) {

            $(".sidebar-toggle").trigger("click");
        }

    }
})


/*
根据select类和选择的状态值, 页面回显
*/
if (typeof statusSelectMultiBox == "object" && statusSelectMultiBox != null) {
    $('.statusSelectMultiBox option').each(function(){
        if(statusSelectMultiBox.indexOf($(this).val()) >= 0){
            $(this).prop("selected",true)
        }
    });
}

if (typeof orderTypeMultiBox == "object" && orderTypeMultiBox != null) {
    $('.orderTypeMultiBox option').each(function(){
        if(orderTypeMultiBox.indexOf($(this).val()) >= 0){
            $(this).prop("selected",true)
        }
    });
}

function getParam(name){
    var sValue=location.search.match(new RegExp("(\\?|&)" + name + "=([^&]*)(&|$)"));
    return sValue?sValue[2]:sValue
}

function popUpModal4GeneratePaymentCode(generateBtnIdPre, repayBalancePre, generateTxtPre, generateUrl){
    $("[id^="+generateBtnIdPre+"]").each(function(){

        $(this).click(function(){

            id_str = $(this).attr("id");
            id_arr = id_str.split("_");
            orderId = id_arr[id_arr.length-1];
            amount = $("#"+repayBalancePre+orderId).text();
            //alert(amount);

            $.ajax({
                url:generateUrl,
                data:{order_id: orderId, balance: amount},
                dataType:'json',
                cache:false,
                type:'post',
                error: function () {

                },
                success:function(result) {
                    if(result.error != ""){
                        alert(result.error)
                    }
                    console.log(result);
                    //$("#remindDialog").modal({})
                    //alert("#repay_balance_"+orderId);
                    //$("#repay_balance_"+orderId).innerHTML= result.amount;

                },
                complete:function() {
                    window.location.reload()

                },
                beforeSend:function() {
                    $("#"+generateTxtPre+orderId).html("<span style='color:red'>Processing, please wait...</span>")
                }
            });
        });
    });
}
//add version

function changePaymentCodeRepayMoney(changeBtnPre, repayBalancePre){
    $("[id^="+changeBtnPre+"]").each(function(){
        $(this).click(function(){
            id_str = $(this).attr("id");
            id_arr = id_str.split("_");
            orderId = id_arr[id_arr.length-1];

            repayBalanceObj = $("#"+repayBalancePre+orderId)
            //amount = repayBalanceObj.attr("balance");
            amount = repayBalanceObj.text();
            repayBalanceObj.html("<input type='number' value="+amount+" />");

            repayInput = repayBalanceObj.children()[0]
            $(repayInput).focus();

            $(repayInput).blur(function(){
                checkValue = $(this).val() - 0;
                if (checkValue <= 0){
                    alert("value error");
                    return false;
                }
                repayBalanceObj.html(checkValue)
            });
        });
    });
}
