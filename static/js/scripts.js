$(function () {
    $(".editBtn").click(function () {
        var editId = $(this).attr("id");
        var challengeId = editId.split("-")[1];
        var editFormId = "editForm-" + challengeId;
        $("#" + editFormId).toggle(500);
    });

    $(".btnOK").click(function () {
        var editId = $(this).attr("id");
        var challengeId = editId.split("-")[1];
        var formId = "editChallengeForm-" + challengeId;
        var frm = $.getElementById('#'+formId);

        frm.submit(function (e) {


            var url = frm.attr("action"); // the script where you handle the form input.
            alert(url);

            $.ajax({
                type: "POST",
                url: url,
                data: frm.serialize(), // serializes the form's elements.
                success: function (data) {
                    var editFormId = "editForm-" + challengeId;
                    $("#" + editFormId).toggle(500);
                    alert(data)
                }

            });
            e.preventDefault(); // avoid to execute the actual submit of the form.
        });
    });

    $('.parallax').parallax();

    $( "#target_time" ).datepicker({
        changeMonth: true,
        changeYear: true,
        numberOfMonths: 1,
        minDate: 0});

    $('.progress-bar').each(function(){
    });
});
