
function fillRepositorySummaries() {
    $.getJSON('/repositories', function(data) {
        $.each(data, function(index, obj) {

            $.getJSON('/tip/master/' + obj['RelativePath'], function(info) {
                $("#" + obj['Name'] + "-sha").text(info['SHA']);
                $("#" + obj['Name'] + "-author").text(info['Author']);
                $("#" + obj['Name'] + "-date").text(info['Date']);
            });

        });
    });
}

$(document).ready(function() {
    fillRepositorySummaries();
});
