let hasNewMergeRequests = function () {
    let statuses = $(".mr-status");
    for (let i=0;i<statuses.length;i++) {
        if ($(statuses[i]).html().trim() === "NEW") {
            return true;
        }
    }

    return false;
}

/* ereg-name */

let app = Vue.createApp({
    mounted() {
    },
    data() {
        return {
            externalRegPopupShow: false,
            backdropShow: false,
            internalRegistryReg: true,
            externalSystemType: "internal-registry",
            removeExternalRegPopupShow: false,
            systemToDelete: '',
            systemToDisable: '',
            mrError: false,
            externalKey: false,
            systemToShowKey: '',
            keyValue: '******',
            currentExternalKeyValue: '',
            accessGrantError: false,
        }
    },
    methods: {
        hideMrError(e) {
            this.mrError = false;
            this.backdropShow = false;

            $("body").css("overflow", "scroll");
            e.preventDefault();
        },
        showMrError(e) {
            this.mrError = true;
            this.backdropShow = true;

            $("body").css("overflow", "hidden");
            e.preventDefault();
            window.scrollTo(0, 0);
        },
        setInternalRegistryReg() {
            this.internalRegistryReg = true;
            this.externalSystemType = "internal-registry";
        },
        setExternalSystem() {
            this.internalRegistryReg = false;
            this.externalSystemType = "external-system";
        },
        showExternalReg(e) {
            if (hasNewMergeRequests()) {
                this.showMrError(e);
                return;
            }

            $("body").css("overflow", "hidden");
            e.preventDefault();
            window.scrollTo(0, 0);

            this.externalRegPopupShow = true;
            this.backdropShow = true;
        },
        hideExternalReg(e) {
            e.preventDefault();
            $("body").css("overflow", "scroll");
            this.externalRegPopupShow = false;
            this.internalRegistryReg = true;
            this.backdropShow = false;
            this.accessGrantError = false;
        },
        disabledLink(e) {
            e.preventDefault();
            return false;
        },
        disableExternalReg(name, e) {
            e.preventDefault();

            if (hasNewMergeRequests()) {
                this.showMrError(e);
                return;
            }

            this.systemToDisable = name;
            $("#disable-form-value").val(name);
            $("#disable-form").submit();

        },
        removeExternalReg(name, e) {
            if (hasNewMergeRequests()) {
                this.showMrError(e);
                return;
            }

            this.systemToDelete = name;
            e.preventDefault();
            this.backdropShow = true;
            this.removeExternalRegPopupShow = true;
            window.scrollTo(0, 0);
            $("body").css("overflow", "hidden");
        },
        hideRemoveExternalReg(e) {
            e.preventDefault();
            this.backdropShow = false;
            this.removeExternalRegPopupShow = false;
            $("body").css("overflow", "scroll");

        },
        showExternalKeyValue(e) {
            if (this.keyValue === '******') {
                this.keyValue = this.currentExternalKeyValue;
            } else {
                this.keyValue = '******'
            }

            e.preventDefault();
        },
        showExternalKey(name, keyValue, e) {
            e.preventDefault();
            this.backdropShow = true;
            this.externalKey = true;
            this.systemToShowKey = name;
            this.currentExternalKeyValue = keyValue;
        },
        hideExternalKey(e) {
            e.preventDefault();
            this.backdropShow = false;
            this.externalKey = false;
            this.keyValue = '******'
        },
        addExternalReg(e) {
            let names = $(".ereg-name");
            let inputName = $("#ex-system").val().trim();
            for (let i=0;i<names.length;i++) {
                if ($(names[i]).html().trim() === inputName) {
                    this.accessGrantError = `Доступ з таким ім'ям системи/або платформи "${inputName}" вже існує, оберіть інше ім'я`;
                    e.preventDefault();
                }
            }
        },
    }
});

app.config.compilerOptions.delimiters = ['[[', ']]']
app.mount('#registry-view')