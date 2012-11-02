
function fillRepositorySummaries() {
    $.getJSON('/repositories', function(data) {
        $.each(data, function(index, obj) {

            $.getJSON('/tip/master/' + obj['RelativePath'], function(info) {
                var name = obj['RelativePath'];
                name = name.replace("/", "_");
                name = name.replace(".", "_");
                $("#" + name + "-sha").text(info['SHA']);
                $("#" + name + "-author").text(info['Author']);
                $("#" + name + "-date").text(info['Date']);
            });

        });
    });
}

function searchFilter() {
    var regexp = $("input#search").val();
    $("tr").each(function(index, obj) {
        var name = $(obj).attr('id');
        if (name.match(regexp) === null) {
            $(obj).hide();
        } else {
            $(obj).show();
        }
    });
}


$(document).ready(function() {
    $("input#search").keyup(function () {
        setTimeout(searchFilter, 50);
    });

    setTimeout(fillRepositorySummaries, 5);
});
