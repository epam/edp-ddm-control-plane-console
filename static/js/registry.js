$(function (){
    let registryName;
    let registryInput = $("#registry-name");

    let hidePopup = function (){
        $(".popup-backdrop").hide();
        $(".popup-window").hide();
        registryInput.val('');
    };

    let showPopup = function (){
        $(".popup-backdrop").show();
        $(".popup-window").show();
        registryInput.val('');
    };

    $(".popup-close").click(function (){
        hidePopup();

        return false;
    });

    $("#delete-cancel").click(function () {
        hidePopup();
        return false;
    })

    registryInput.val('');
    registryInput.keyup(function (e) {
        let popupFooter = $(".popup-footer");
        popupFooter.removeClass('active');
        if (registryName.toString() === $(e.currentTarget).val()) {
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

$(document).ready(function() {
    $("#registry-table").DataTable({
        ordering: true,
        paging: true,
        columnDefs: [
            { orderable: false, targets: 0 },
            { orderable: false, targets: 4 },
            { orderable: false, targets: 5 }
        ],
        order: [[3, 'desc']],
        language: {
            "processing": "Зачекайте...",
            "lengthMenu": "Показати _MENU_ записів",
            "zeroRecords": "Записи відсутні.",
            "info": "Записи з _START_ по _END_ із _TOTAL_ записів",
            "infoEmpty": "Записи з 0 по 0 із 0 записів",
            "infoFiltered": "(відфільтровано з _MAX_ записів)",
            "search": "Пошук:",
            "paginate": {
                "first": "Перша",
                "previous": "Попередня",
                "next": "Наступна",
                "last": "Остання"
            },
            "aria": {
                "sortAscending": ": активувати для сортування стовпців за зростанням",
                "sortDescending": ": активувати для сортування стовпців за спаданням"
            }
        }
    });
});