let app = Vue.createApp({
    mounted() {
        console.log('view registry mounted');
    },
    data() {
        return {
            externalRegPopupShow: false,
            internalRegistryReg: true,
            externalSystemType: "internal-registry",
        }
    },
    methods: {
        setInternalRegistryReg() {
            this.internalRegistryReg = true;
            this.externalSystemType = "internal-registry";
        },
        setExternalSystem() {
            this.internalRegistryReg = false;
            this.externalSystemType = "external-system";
        },
        showExternalReg(e) {
            $("body").css("overflow", "hidden");
            e.preventDefault();
            window.scrollTo(0, 0);
            this.externalRegPopupShow = true;
        },
        hideExternalReg() {
            $("body").css("overflow", "scroll");
            this.externalRegPopupShow = false;
            this.internalRegistryReg = true;
        },
        addExternalReg() {

        },
    }
});

app.config.compilerOptions.delimiters = ['[[', ']]']
app.mount('#registry-view')