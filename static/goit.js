
function fillRepositorySummaries() {
    $("tr").each(function(index, obj) {
        var name = $(obj).attr('id');
        var path = $(obj).attr('relativePath');
        $.getJSON('/tip/master/' + path, function(info) {
            $("#" + name + "-sha").text(info['SHA']);
            $("#" + name + "-author").text(info['Author']);
            $("#" + name + "-date").text(info['Date']);
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
