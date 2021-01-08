$(function (){
    let registryName;

    let hidePopup = function (){
        $(".popup-backdrop").hide();
        $(".popup-window").hide();
    };

    let showPopup = function (){
        $(".popup-backdrop").show();
        $(".popup-window").show();
    };

    $(".popup-close").click(function (){
        hidePopup();

        return false;
    });

    $("#delete-cancel").click(function () {
        hidePopup();
        return false;
    })

    let registryInput = $("#registry-name");

    registryInput.val('');
    registryInput.keyup(function (e) {
        let popupFooter = $(".popup-footer");
        popupFooter.removeClass('active');
        if (registryName === $(e.currentTarget).val()) {
            popupFooter.addClass('active');
        }
    });

    $(".delete-registry").click(function (e){
        registryName = $(e.currentTarget).data('name');
        $("#delete-name").html(registryName);

        showPopup();
    });

    $("#delete-form").submit(function () {
        return registryName === registryInput.val()
    });
});