
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

function searchFilter() {
    var regexp = $("input#search").val();
        console.log(regexp);
    $("tr").each(function(index, obj) {
        var name = $(obj).attr('id');
        console.log(name);
        if (name.match(regexp) === null) {
            $(obj).hide();
        } else {
            $(obj).show();
        }
    });
}


$(document).ready(function() {
    fillRepositorySummaries();
    $("input#search").keyup(searchFilter);
});
