$(function () {
    var last_sort_id = ""
    var last_sort_type = ""

    $(document).ready(function () {
        last_sort_id = getParam("field")
        last_sort_type = getParam("sort")

        var classname = ""
        if (last_sort_type == "DESC") {
            classname = "glyphicon-chevron-down"
        } else if (last_sort_type == "ASC") {
            classname = "glyphicon-chevron-up"
        } else {
            return
        }

        var lastEle = document.getElementById("i" + last_sort_id)
        if (null != lastEle) {
            lastEle.classList.remove("glyphicon-resize-vertical")
            lastEle.classList.add(classname)
        }
    })

    $('.th_sort').click(function() {
        var id = $(this).attr('id')

        if (last_sort_id == id){
            if (last_sort_type == "ASC"){
                last_sort_type = "DESC"
            } else {
                last_sort_type = "ASC"
            }
        } else {
            last_sort_id = id
            last_sort_type = "DESC"
        }

        var url = location.href
        if (-1 != url.indexOf('field')){
            url = url.replace(new RegExp("field" + "=([^&]*)(&|$)"), '')
            url = url.replace(new RegExp("sort" + "=([^&]*)(&|$)"), '')
        }

        if (-1 != url.indexOf('?')) {
            if ('&' == url.substr(url.length - 1, 1)) {
                url = url + 'field=' + last_sort_id + '&sort=' + last_sort_type

            } else {
                url = url + '&field=' + last_sort_id + '&sort=' + last_sort_type
            }
        } else {
            url =  url + '?field=' + last_sort_id + '&sort=' + last_sort_type
        }

        window.location = url

    })
})


