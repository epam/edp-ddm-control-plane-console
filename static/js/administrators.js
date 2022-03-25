let app = Vue.createApp({
    mounted() {
        // console.log(this.adminsValue);
    },
    data() {
        return {
            adminsValue: '',
            message: 'Hello Vue!',
            adminPopupShow: false,
            admins: [],
            editAdmin: {
                username: "",
                firstName: "",
                lastName: "",
                email: "",
                tmpPassword: ""
            },
            requiredError: false,
            emailFormatError: false,
            usernameFormatError: false,
            adminsLoaded: false,
        }
    },
    methods: {
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
            this.usernameFormatError = false;

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

            if (!String(this.editAdmin.username).match(/^[a-z0-9\-]{3,}$/)) {
                this.usernameFormatError = true;
                return;
            }

            $("body").css("overflow", "scroll");
            this.adminPopupShow = false;

            this.admins.push({
                username: this.editAdmin.username,
                email: this.editAdmin.email,
                firstName: this.editAdmin.firstName,
                lastName: this.editAdmin.lastName,
                tmpPassword: this.editAdmin.tmpPassword
            });

            this.editAdmin = {
                username: "",
                firstName: "",
                lastName: "",
                email: "",
                tmpPassword: ""
            };

            this.adminsValue = JSON.stringify(this.admins);

        }
    }
})

app.config.compilerOptions.delimiters = ['[[', ']]']
app.mount('#registry-form')