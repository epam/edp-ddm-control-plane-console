let app = Vue.createApp({
    mounted() {
        if (this.$refs.hasOwnProperty('smtpServerTypeSelected')) {
            let selectedSMTP = this.$refs.smtpServerTypeSelected.value
            if (selectedSMTP === "") {
                selectedSMTP = "platform-mail-server"
            }
            this.smtpServerType = selectedSMTP;
        }

        if (this.$refs.hasOwnProperty('smtpEditConfig')) {
            let smtpConfig = JSON.parse(this.$refs.smtpEditConfig.value);
            if (smtpConfig['type'] === 'external') {
                this.smtpServerType = 'external-mail-server';
                this.externalSMTPOpts = smtpConfig;
                this.externalSMTPOpts['port'] = smtpConfig['port'].toString();
            } else {
                this.smtpServerType = 'platform-mail-server';
            }
        }

        if (this.$refs.hasOwnProperty('cidrEditConfig')) {
            let cidrConfig = JSON.parse(this.$refs.cidrEditConfig.value);

            if (cidrConfig.hasOwnProperty('citizen')) {
                this.citizenCIDR = cidrConfig.citizen;
                this.citizenCIDRValue.value = JSON.stringify(this.citizenCIDR);
            }

            if (cidrConfig.hasOwnProperty('officer')) {
                this.officerCIDR = cidrConfig.officer;
                this.officerCIDRValue.value = JSON.stringify(this.officerCIDR);
            }

            if (cidrConfig.hasOwnProperty('admin')) {
                this.adminCIDR = cidrConfig.admin;
                this.adminCIDRValue.value = JSON.stringify(this.adminCIDR);
            }
        }
    },
    data() {
        return {
            officerCIDRValue: { value: '' },
            officerCIDR: [],
            citizenCIDRValue: { value: '' },
            citizenCIDR: [],
            adminCIDRValue: { value: '' },
            adminCIDR: [],
            adminsValue: '',
            currentCIDR: [],
            currentCIDRValue: '',
            cidrFormatError: false,
            adminPopupShow: false,
            cidrPopupShow: false,
            admins: [],
            editCIDR: '',
            editAdmin: {
                firstName: "",
                lastName: "",
                email: "",
                tmpPassword: ""
            },
            requiredError: false,
            emailFormatError: false,
            adminsLoaded: false,
            adminsError: false,
            smtpServerType: null,
            mailServerOpts: '',
            externalSMTPOpts: {
                host: '',
                port: '587',
                address: '',
                password: ''
            }
        }
    },
    methods: {
        registryFormSubmit(e) {
            if (this.admins.length === 0) {
                this.adminsError = true;
                e.preventDefault();

                let element = this.$refs['admins'];
                let top = element.offsetTop;

                window.scrollTo(0, top);
            }

            this.mailServerOpts = JSON.stringify(this.externalSMTPOpts);
        },
        loadAdmins(admins) {
            if (!this.adminsLoaded) {
                this.admins = JSON.parse(admins);
                this.adminsValue = JSON.stringify(this.admins);
                this.adminsLoaded = true;
            }
        },
        showAdminForm() {
            this.emailFormatError = false;
            this.requiredError = false;
            this.adminPopupShow = true;
            $("body").css("overflow", "hidden");
        },
        showCIDRForm(cidr, value) {
            this.cidrPopupShow = true;
            $("body").css("overflow", "hidden");

            this.editCIDR = '';
            this.currentCIDR = cidr;
            this.currentCIDRValue = value;
            this.cidrFormatError = false;
        },
        hideCIDRForm() {
            this.cidrPopupShow = false;
            $("body").css("overflow", "scroll");
        },
        createCIDR(e) {
            e.preventDefault();
            if (!String(this.editCIDR).toLowerCase().match(/^([01]?\d\d?|2[0-4]\d|25[0-5])(?:\.(?:[01]?\d\d?|2[0-4]\d|25[0-5])){3}(?:\/[0-2]\d|\/3[0-2])?$/)) {
                this.cidrFormatError = true;
                return;
            }

            this.currentCIDR.push(this.editCIDR);
            this.currentCIDRValue.value = JSON.stringify(this.currentCIDR);
            this.hideCIDRForm();

        },
        deleteCIDR(c, cidr, value, e) {
            e.preventDefault();

            for (let v in cidr) {
                if (cidr[v] === c) {
                    cidr.splice(v, 1);
                    break;
                }
            }

            value = JSON.stringify(cidr);
        },
        hideAdminForm() {
            this.adminPopupShow = false;
            $("body").css("overflow", "scroll");
        },
        deleteAdmin: function (e) {
            e.preventDefault();
            let email = e.currentTarget.getAttribute('email');

            for (let v in this.admins) {
                if (this.admins[v].email === email) {
                    this.admins.splice(v, 1);
                    break;
                }
            }
            this.adminsValue = JSON.stringify(this.admins);
        },
        createAdmin: function (e) {
            this.requiredError = false;
            this.emailFormatError = false;

            e.preventDefault();
            for (let v in this.editAdmin) {
                if (this.editAdmin[v] === "") {
                    this.requiredError = true;
                    return;
                }
            }

            if (!String(this.editAdmin.email)
                .toLowerCase()
                .match(
                    /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
                )) {
                this.emailFormatError = true;
                return;
            }

            $("body").css("overflow", "scroll");
            this.adminPopupShow = false;

            this.admins.push({
                email: this.editAdmin.email,
                firstName: this.editAdmin.firstName,
                lastName: this.editAdmin.lastName,
                tmpPassword: this.editAdmin.tmpPassword
            });

            this.editAdmin = {
                firstName: "",
                lastName: "",
                email: "",
                tmpPassword: ""
            };

            this.adminsValue = JSON.stringify(this.admins);
        }
    }
})

app.config.compilerOptions.delimiters = ['[[', ']]'];
app.mount('#registry-form');

let editKeyChecked = function(){
    let checked = $("#edit-key").prop('checked'),
        keyBlock = $("#key-block");
    keyBlock.find("input").prop('disabled', !checked);

    if (checked) {
        keyBlock.show();
        $("#key-device-type").change();
    } else {
        keyBlock.hide();
    }
};

$(function () {
    let editKey = $("#edit-key");
    editKey.change(editKeyChecked);
    editKey.prop("checked", false);
    editKeyChecked();
});