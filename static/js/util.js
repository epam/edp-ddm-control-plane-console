function getUrlParameter(sParam) {
    let sPageURL = window.location.search.substring(1),
        sURLVariables = sPageURL.split('&'),
        sParameterName,
        i;

    for (i = 0; i < sURLVariables.length; i++) {
        sParameterName = sURLVariables[i].split('=');

        if (sParameterName[0] === sParam) {
            return sParameterName[1] === undefined ? true : decodeURIComponent(sParameterName[1]);
        }
    }
}

function _sendPostRequest(async, url, data, token, successCallback, failCallback) {
    $.ajax({
        url: url,
        contentType: "application/json",
        type: "POST",
        data: JSON.stringify(data),
        async: async,
        headers: {
            'X-Csrftoken': token
        },
        success: function (resp) {
            successCallback(resp);
        },
        error: function (resp) {
            failCallback(resp);
        }
    });
}

function htmlEncode(value){
    return $('<div/>').text(value).html();
}

function isFieldValid(elementToValidate, regex) {
    let check = function (value) {
        return regex.test(value);
    };

    return !(!elementToValidate.val() || !check(elementToValidate.val()));
}

function isCodebaseSiteFieldValid(elementToValidate, regex) {
    return regex.test(elementToValidate.val());
}

function blockIsNotValid($block) {
    $block.find('.card-header')
        .addClass('invalid')
        .removeClass('success')
        .addClass('error');
}

function blockIsValid($block) {
    $block.find('.card-header')
        .removeClass('invalid')
        .addClass('success')
        .removeClass('error');
}


function createConfirmTable(formName) {

    let $formData = $(formName).serializeArray();

    let getValue = function (name) {
        let record = $formData.find(x => x.name === name);
        return record !== undefined ? record.value : "";
    };

    let isFound = function (name) {
        return $formData.find(x => x.name === name)
    };

    let addBlock = function (qwery, name, block) {
        let result = "";
        let show = typeof query === "boolean" ? qwery : typeof qwery === "string" ? isFound(qwery) : true;
        if (show) {
            if (name) {
                result += '<tr><td class="font-weight-bold text-center" colspan="2">' + name + '</td></tr>';
            }
            $.each(block, function (key, property) {
                let value = getValue(property);
                value = typeof property === 'boolean' ? (property ? "&#10004;" : "&#10008;") : htmlEncode(getValue(property));
                if (value) {
                    if (key === 'Start Versioning From') {
                        result += '<tr><td>' + key + '</td><td>' + value + '-SNAPSHOT' + '</td></tr>';
                    } else {
                        result += '<tr><td>' + key + '</td><td>' + value + '</td></tr>';
                    }
                }
            });
            $(result).appendTo($("#window-table"));
        }
    };

    $('<tbody class="window-table-body">').appendTo($("#window-table"));

    addBlock(null, null,
        {
            'Name': 'appName',
            'Description': 'description',
            'Code language': 'appLang',
            'Framework': 'framework',
            'Build tool': 'buildTool',
            'Integration with VCS is enabled': $('.vcs-block').length !== 0,
            'Multi-module project': isFound('isMultiModule')
        });

    addBlock(null, "CODEBASE",
        {
            'Integration method': 'strategy'
        });

    let advancedBlock = {
        'Job Provisioner': 'jobProvisioning',
        'Jenkins Slave': 'jenkinsSlave',
        'Deployment Script': 'deploymentScript',
        'Versioning Type': 'versioningType',
        'Commit Message Pattern': 'commitMessagePattern',
        'Ticket Name Pattern': 'ticketNamePattern',
        'CI tool': 'ciTool'
    };

    if ($('#versioningType').val() === 'edp') {
        advancedBlock['Start Versioning From'] = 'startVersioningFrom'
    } else {
        delete advancedBlock['Start Versioning From'];
    }

    if ($('#jiraServerToggle').is(':checked')) {
        advancedBlock['Jira Server'] = 'jiraServer'
    } else {
        delete advancedBlock['Jira Server'];
    }

    addBlock(null, "ADVANCED SETTINGS", advancedBlock);

    if (!isFound('strategy') || getValue('strategy') === "clone") {
        addBlock(
            null, null,
            {'Repository': 'gitRepoUrl'});
        addBlock('isRepoPrivate', null,
            {'Login': 'repoLogin'});
    }

    addBlock($('.vcs-block').length !== 0, null,
        {'VCS Login': 'vcsLogin'});

    $(addBlock('needRoute', 'EXPOSING SERVICE INFO', {
        'Exposing service name': 'routeSite',
        'Exposing service path': 'routePath'
    }));

    $(addBlock('testReportFramework', 'REPORT FRAMEWORK', {
        'Autotest Report Framework': 'testReportFramework'
    }));

    $(addBlock('needDb', 'DATABASE', {
        'Database': 'database',
        'Version': 'dbVersion',
        'Capacity': 'dbCapacity',
        'Persistent storage': 'dbPersistentStorage'
    }));

}

function _sendGetRequest(async, url, successCallback, failCallback) {
    $.ajax({
        url: url,
        contentType: "application/json",
        async: async,
        success: function (resp) {
            successCallback(resp);
        },
        error: function (resp) {
            failCallback(resp);
        },
    });
}
