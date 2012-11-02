
function fillRepositorySummaries() {
    $.getJSON('/repositories', function(data) {
        $.each(data, function(index, obj) {

            $.getJSON('/tip/master/' + obj['RelativePath'], function(info) {
                var name = obj['Name'];
                name = name.replace("/", "_");
                name = name.replace(".", "_");
                $("#" + name + "-sha").text(info['SHA']);
                $("#" + name + "-author").text(info['Author']);
                $("#" + name + "-date").text(info['Date']);
            });

        });
    });
}

$(document).ready(function() {
    fillRepositorySummaries();
});
