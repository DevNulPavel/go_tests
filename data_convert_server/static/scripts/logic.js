function disableAll() {
    radios = document.getElementsByName('convertType');
    for (var i = 0; i< radios.length;  i++) {
        radios[i].disabled = true;
        radios[i].checked = false;
    }
    convertButton = document.getElementById("convertButton")
    convertButton.disabled = true
}

function enableButton(){
    convertButton = document.getElementById("convertButton")
    convertButton.disabled = !validInputType
}

function showErrorText() {
    errorMesage = document.getElementById("errorMesage")
    errorMesage.hidden = false
}

function hideErrorText() {
    errorMesage = document.getElementById("errorMesage")
    errorMesage.hidden = true
}

function onFileSelected() {
    // var filename = $('input[type=file]').val().split('\\').pop();
    var files = $('#fileChoose').prop("files")
    if ((files == null) || (files.length == 0)){
        disableAll()
        showErrorText()
        return
    }

    extentions = []
    for (var i = 0; i < files.length; i++){
        fileExt = files[i].name.split('.').pop();
        fileExt = fileExt.toLowerCase();
        extentions.push(fileExt)
        console.log(fileExt)
    }

    if (extentions.length == 0){
        disableAll()
        showErrorText()
        return
    }

    normalImagesExt = ["jpeg", "jpg", "png"]
    webpImagesExt = ["webp"]
    audioExt = ["wav", "mp3", "ogg"]
    videoExt = ["avi", "mp4", "mkv"]

    type = -1
    for (var i = 0; i < extentions.length; i++) {
        if ($.inArray(extentions[i], normalImagesExt) >= 0){
            testTypeVal = 1
            if(type < 0){
                type = testTypeVal
            } else if (type != testTypeVal){
                disableAll()
                showErrorText()
                return
            }
        }else if ($.inArray(extentions[i], webpImagesExt) >= 0){
            testTypeVal = 2
            if(type < 0){
                type = testTypeVal
            } else if (type != testTypeVal){
                disableAll()
                showErrorText()
                return
            }
        }else if ($.inArray(extentions[i], audioExt) >= 0){
            testTypeVal = 3
            if(type < 0){
                type = testTypeVal
            } else if (type != testTypeVal){
                disableAll()
                showErrorText()
                return
            }
        }else if ($.inArray(extentions[i], videoExt) >= 0){
            testTypeVal = 4
            if(type < 0){
                type = testTypeVal
            } else if (type != testTypeVal){
                disableAll()
                showErrorText()
                return
            }
        }else {
            disableAll()
            showErrorText()
            return
        }
    }

    validValues = []

    if ($.inArray(extentions[0], normalImagesExt) >= 0){
        validValues = ["pvr", "pvrgz16", "pvrgz32", "webp"]
    }else if ($.inArray(extentions[0], webpImagesExt) >= 0){
        validValues = ["png"]
    }else if ($.inArray(extentions[0], audioExt) >= 0){
        validValues = ["m4a", "ogg"]
    }else if ($.inArray(extentions[0], videoExt) >= 0){
        validValues = ["mp4", "webm"]
    }

    validInputType = false
    anyoneSelected = false
    radios = document.getElementsByName('convertType');
    for (var i = 0; i< radios.length;  i++){
        value = radios[i].value
        if($.inArray(value, validValues) >= 0){
            if(anyoneSelected == false){
                radios[i].checked = true
                anyoneSelected = true
            }
            radios[i].disabled = false;
            validInputType = true
        }else{
            radios[i].checked = false
            radios[i].disabled = true;
        }
    }

    enableButton()
    hideErrorText()
}