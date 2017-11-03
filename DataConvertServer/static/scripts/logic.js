function onFileSelected() {
    var filename = $('input[type=file]').val().split('\\').pop();
    fileExt = filename.split('.').pop();
    fileExt = fileExt.toLowerCase();
    console.log(fileExt)

    validValues = []

    imagesExt = ["jpeg", "png"]
    audioExt = ["wav", "mp3", "ogg"]
    if ($.inArray(fileExt, imagesExt) >= 0){
        validValues = ["pvr", "pvrgz16", "pvrgz32"]
    }else if ($.inArray(fileExt, audioExt) >= 0){
        validValues = ["m4a", "ogg"]
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

    convertButton = document.getElementById("convertButton")
    convertButton.disabled = !validInputType
}