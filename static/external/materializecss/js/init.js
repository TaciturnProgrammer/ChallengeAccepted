$(document).ready(function () {
    $('.parallax').parallax();
    $('.modal-trigger').leanModal();

    $('.datepicker').pickadate({
        min: "tomorrow",
        selectMonths: true, // Creates a dropdown to control month
        selectYears: 2 // Creates a dropdown of 15 years to control year
    });

    var difference = $('.hiddenDifference');
    difference.each(function () {
        if ($(this).text() > 0) {
            $(this).parent().removeClass('green');
            $(this).parent().addClass('red darken-4');
        }
    });


});
