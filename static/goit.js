
var regexp;
var repositories = new Array();

function fillRepositorySummaries() {
    $("tr").each(function(index, obj) {
        obj = $(obj);
        repositories.push(obj);
        var name = obj.attr('id');
        var path = obj.attr('relativePath');
        $.getJSON('/tip/master/' + path, function(info) {
            $("#" + name + "-sha").text(info['SHA']);
            $("#" + name + "-author").text(info['Author']);
            $("#" + name + "-date").text(info['Date']);
            $("#" + name + "-subject").text(info['Subject']);
        });
    });
}

function searchFilter() {
    for (var i=0; i < repositories.length; i++) {
        var obj = repositories[i];
        var name = obj.attr('id');
        if (regexp.test(name) === false) {
            obj.css({'display': 'none'});
//            obj.hide();
        } else {
            obj.css({'display': 'table-row'});
//            obj.show();
        }
    }
}


$(document).ready(function() {
    var search_input = $("input#search")
    search_input.keyup(function () {
        regexp = new RegExp(search_input.val(), "i");
        setTimeout(searchFilter, 50);
    });

    setTimeout(fillRepositorySummaries, 5);
});
