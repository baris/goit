goit = window.goit || {};

goit.init = function () {
    goit.Globals = {};
    goit.Globals.searchRegexp = "";
    goit.Globals.fetchedRepositories = false;
    goit.Globals.repositories = new Array();
};


goit.idFromPath = function (path) {
    return path.replace(/-/g, '_').replace(/\//g, '_').replace(/\./g, '_');
};

goit.handleLoginError = function(d) {
    if (d.hasOwnProperty('login_error') && window.location.pathname !== "/login.html") {
        window.location = "/login.html";
    }
};

goit.fillRepositories = function () {
    if (goit.Globals.fetchedRepositories === true) {
        return;
    }
    $.getJSON('/repositories/', function(repos) {
        goit.handleLoginError(repos);

        for (var i = 0; i < repos.length; i++) {
            var txt = "";
            var repo = repos[i];
            var id = goit.idFromPath(repo.RelativePath);
	    obj = $("<tr id="+id+" relativePath='"+repo.RelativePath+"'>"+
                    "<td class=repo_name><a href=/repository.html#"+repo.RelativePath+">"+repo.RelativePath+"</a></td>"+
                    "<td class=repo_gitweb><a href="+repo.GitwebUrl+">Gitweb Page</a></td>"+
                    "<td class=repo_sha id="+id+"-sha></td>"+
                    "<td class=repo_author id="+id+"-author></td>"+
                    "<td class=repo_subject id="+id+"-subject></td>"+
                    "<td class=repo_date id="+id+"-date></td>"+
                    "</tr>");
            $('#repositories-table').append(obj);
        }
        goit.Globals.fetchedRepositories = true;
        $("tr").each(function (idx, obj) {
            goit.Globals.repositories.push($(obj));
        });
    });
};


goit.fillRepositorySummaries = function () {
    for (var i = 0; i < goit.Globals.repositories.length; i++) {
        obj = goit.Globals.repositories[i];
        var path = obj.attr('relativePath');
        $.getJSON('/tip/master/' + path, function(ret) {
            goit.handleLoginError(ret);

            var repo = ret[0]; var info = ret[1];
            var id = goit.idFromPath(repo.RelativePath);
            $("#" + id + "-sha").html(
                $("<a>", {
                    "href": repo.GitwebUrl+";a=commitdiff;h="+info['SHA'],
                    "text": info['SHA'],
                })
            );
            $("#" + id + "-author").html(
                $("<a>", {
                    "href": repo.GitwebUrl+";a=search;s="+info['Author']+";st=author",
                    "text": info['Author'],
                })
            );
            $("#" + id + "-date").text(info['Date']);
            $("#" + id + "-subject").text(info['Subject']);
        });
    }
};


goit.searchFilter = function () {
    for (var i = 0; i < goit.Globals.repositories.length; i++) {
        var obj = goit.Globals.repositories[i];
        var klass = obj.attr('class');
        if (klass == "table-header") {
            continue;
        }

        var name = obj.attr('id');
        if (goit.Globals.searchRegexp.test(name) === false) {
            obj.css({'display': 'none'});
        } else {
            obj.css({'display': 'table-row'});
        }
    }
};


goit.repositoriesReady = function () {
    if (goit.Globals.fetchedRepositories == false) {
        setTimeout(goit.repositoriesReady, 1);
        return;
    }

    var search_input = $("input#search");
    search_input.keyup(function () {
        goit.Globals.searchRegexp = new RegExp(goit.idFromPath(search_input.val()), "i");
        setTimeout(goit.searchFilter, 50);
    });
    goit.fillRepositorySummaries();
};


goit.showRepository = function (repository, limit) {
    $.getJSON('/commits/master/'+limit+'/'+repository, function(ret) {
        goit.handleLoginError(ret);

        var repo = ret[0]; var commits = ret[1];
        var id = goit.idFromPath(repo.RelativePath);
        for (var i = 0; i < commits.length; i++) {
            commit = commits[i];
            obj = $("<tr id="+id+" relativePath='"+repo.RelativePath+"'>"+
                    "<td><a href="+repo.GitwebUrl+";a=commitdiff;h="+commit.SHA+">"+commit.SHA+"</a></td>"+
                    "<td><a href="+repo.GitwebUrl+";a=search;s="+commit.Author+";st=author>"+commit.Author+"</a></td>"+
                    "<td>"+commit.Subject+"</td>"+
                    "<td>"+commit.Date+"</td>"+
                    "</tr>");
            $('#repository-table').append(obj);
        }
    });

    $.getJSON('/heads/'+repository, function(ret) {
        goit.handleLoginError(ret);

        var repo = ret[0];
        var heads = ret[1];
        var id = goit.idFromPath(repo.RelativePath);
        for (var i = 0; i < heads.length; i++) {
            head = heads[i];
            console.log(head);
            obj = $("<tr id="+id+" relativePath='"+repo.RelativePath+"'>"+
	            "<td><a href="+repo.GitwebUrl+";a=shortlog;h=refs/heads/"+head+">"+
	            head+
	            "</a></td>"+
	            "</tr>");
            $('#heads-table').append(obj);
        }
    });
};


$(document).ready(function() {
    goit.init();
    var url = $(location).attr('href').match('.*#(.*)');
    if (url && url[1].length > 0) {
        parts = url[1].match('(.*)~(.*)~(.*)');
        repository = "";
        limit = 10;
        if (parts && parts[1].length > 0 && parts[2].length > 0) {
            repository = parts[1];
            limit = parts[2];
        } else {
            repository = url[1];
        }
        goit.showRepository(repository, limit);
    } else {
        goit.fillRepositories();
        goit.repositoriesReady();
    }
});
