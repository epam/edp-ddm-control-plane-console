<script setup lang="ts">
import Typography from '@/components/common/Typography.vue';
import Banner from '@/components/common/Banner.vue';
import FileField from '@/components/common/FileField.vue';
import isEqual from 'lodash/isEqual';
import cloneDeep from 'lodash/cloneDeep';
</script>

<script lang="ts">
import { defineComponent } from 'vue';
import OsplmIniEditor from '@/utils/osplmIniEditor';

export default defineComponent({
  props: {
    registryAction: String,
    pageDescription: String,
    region: String,
  },
  methods: {
    renderINITemplate() {
      const iniTemplate = document.getElementById("ini-template")?.innerHTML;
      const iniEditor =  new OsplmIniEditor(iniTemplate || '');
      iniEditor.addKey({
        caHost: this.hardwareData.remoteCaHost,
          caPort: this.hardwareData.remoteCaPort,
          caName: '',
          keyHost: this.hardwareData.remoteKeyHost,
          keySn: this.hardwareData.remoteSerialNumber,
          keyMask: this.hardwareData.remoteKeyMask,
      });
      this.hardwareData.iniConfig = iniEditor.toString();
    },
    removeAllowedKey(item: any) {
      const searchIdx = this.allowedKeys.indexOf(item);
      if (searchIdx !== -1) {
        this.allowedKeys.splice(searchIdx, 1);
      }
    },
    addAllowedKey() {
      this.allowedKeys.push({ issuer: "", serial: "", removable: true });
    },
    hardwareDataChanged() {
      this.renderINITemplate();
    },
    onKey6FileSelected(){
      this.key6FileSelected = true;
      this.key6Error = '';
    },
    onKey6FileReset(){
      this.key6FileSelected = false;
    },
    validator() {
      return new Promise<void>((resolve, reject) => {
        if (this.registryAction === "edit" && !this.key6FileSelected && this.isDataChanged) {
          this.changed = false;
          this.beginValidation = false;
          this.key6Error = '';
          resolve();
          return true;
        }
        this.renderINITemplate();
        this.changed = true;
        this.beginValidation = true;
        this.key6Error = '';
        let validationFailed = false;

        for (let i = 0; i < this.allowedKeys.length; i++) {
          if (this.allowedKeys[i].issuer === "" ||
              this.allowedKeys[i].serial === "") {
            validationFailed = true;
          }
        }
        if (this.deviceType === "hardware") {
          for (const key in this.hardwareData) {
            if (this.hardwareData[key as keyof typeof this.hardwareData] === "") {
              validationFailed = true;
            }
          }
        } else {
          for (const key in this.fileData) {
            if (this.fileData[key as keyof typeof this.fileData] === "") {
              validationFailed = true;
            }
          }

          if (!this.key6FileSelected) {
            this.key6Error = this.$t('errors.requiredField');
            validationFailed = true;
          }
        }

        if (validationFailed) {
          reject();
          return false;
        }

        this.beginValidation = false;
        resolve();
        return true;
      });
    },
  },
  data() {
    return {
      changed: false,
      allowedKeys: [{ issuer: "", serial: "", removable: false }],
      beginValidation: false,
      key6Error: '',
      key6FileSelected: false,
      deviceType: "file",
      hardwareData: {
        remoteType: this.$t('components.keyData.text.hardwareDataType'),
        remoteKeyPWD: "",
        remoteCaName: "",
        remoteCaHost: "",
        remoteCaPort: "",
        remoteSerialNumber: "",
        remoteKeyPort: "",
        remoteKeyHost: "",
        remoteKeyMask: "",
        iniConfig: "",
      },
      fileData: {
        signKeyIssuer: "",
        signKeyPWD: "",
      },
      defaultData: {} as Record<string, unknown>,
    };
  },
  mounted() {
    this.defaultData = {
      hardware: cloneDeep(this.$data.hardwareData),
      file: cloneDeep(this.$data.fileData),
      allowedKeys: cloneDeep(this.$data.allowedKeys),
    };
  },
  computed: {
    isDataChanged(): boolean {
      let isEqualDeviceData = isEqual(this.defaultData.file, this.$data.fileData);
      if (this.deviceType === 'hardware') {
        isEqualDeviceData = isEqual(this.defaultData.hardware, this.$data.hardwareData);
      }

      return (
        isEqualDeviceData &&
        isEqual(this.defaultData.allowedKeys, this.$data.allowedKeys)
      );
    },
  }
});
</script>

<style scoped>
    .key-data-page-description {
      margin-bottom: 32px;
    }

    .add-allowed-key-btn {
      display: flex;
      text-decoration: none;
      align-items: baseline;
      margin-bottom: 16px;
    }

    .add-allowed-key-btn:hover {
      text-decoration: none;
    }

    .add-allowed-key-btn div {
      font-family: "Oswald", "MuseoSans", sans-serif;
      font-weight: 400;
      font-size: 18px;
      text-transform: uppercase;
      color: rgba(0, 0, 0, 0.5);
      margin-left: 13px;
    }

</style>

<template>
  <h2>{{ $t('components.keyData.title') }}</h2>
  <div v-if="region === 'ua'">
    <Typography variant="bodyText" class="key-data-page-description">{{ pageDescription }}</Typography>

    <input type="checkbox" style="display: none;" v-model="changed" name="key-data-changed" />
    <div class="rc-form-group">
      <label for="key-device-type">{{ $t('components.keyData.text.mediaType') }}</label>
      <select v-model="deviceType" id="key-device-type"
              name="key-device-type">
        <option value="file">{{ $t('components.keyData.text.fileMedium') }}</option>
        <option value="hardware">{{ $t('components.keyData.text.hardwareMedium') }}</option>
      </select>
    </div>
    <div class="key-type-hardware key-type-section" v-if="deviceType === 'hardware'" v-cloak>
      <div class="rc-form-group"
          :class="{ 'error': hardwareData.remoteType === '' && beginValidation }">
        <label for="remote-type">{{ $t('components.keyData.text.keyType') }}</label>
        <input type="text" @change="hardwareDataChanged" name="remote-type" id="remote-type"
              v-model="hardwareData.remoteType" />
        <span v-if="hardwareData.remoteType === '' && beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group"
          :class="{ 'error': hardwareData.remoteKeyPWD === '' && beginValidation }">
        <label for="remote-key-pwd">{{ $t('components.keyData.text.keyPassword') }}</label>
        <input @change="hardwareDataChanged" type="password" id="remote-key-pwd" name="remote-key-pwd"
              v-model="hardwareData.remoteKeyPWD" />
        <span v-if="hardwareData.remoteKeyPWD === '' && beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group"
          :class="{ 'error': hardwareData.remoteCaName === '' && beginValidation }">
        <label for="remote-ca-name">{{ $t('components.keyData.text.nameOfAcsk') }}</label>
        <input @change="hardwareDataChanged" type="text" id="remote-ca-name" name="remote-ca-name"
              v-model="hardwareData.remoteCaName" />
        <span v-if="hardwareData.remoteCaName === '' && beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group"
          :class="{ 'error': hardwareData.remoteCaHost === '' && beginValidation }">
        <label for="remote-ca-host">{{ $t('components.keyData.text.hostOfAcsk') }}</label>
        <input @change="hardwareDataChanged" type="text" id="remote-ca-host" name="remote-ca-host"
              v-model="hardwareData.remoteCaHost" />
        <span v-if="hardwareData.remoteCaHost === '' && beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group"
          :class="{ 'error': hardwareData.remoteCaPort === '' && beginValidation }">
        <label for="remote-ca-port">{{ $t('components.keyData.text.portOfAcsk') }}</label>
        <input @change="hardwareDataChanged" type="number" id="remote-ca-port" name="remote-ca-port"
              v-model="hardwareData.remoteCaPort" />
        <span v-if="hardwareData.remoteCaPort === '' && beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group"
          :class="{ 'error': hardwareData.remoteSerialNumber === '' && beginValidation }">
        <label for="remote-serial-number">{{ $t('components.keyData.text.deviceSerialNumber') }}</label>
        <input @change="hardwareDataChanged" type="text" id="remote-serial-number" name="remote-serial-number"
              v-model="hardwareData.remoteSerialNumber" />
        <span
            v-if="hardwareData.remoteSerialNumber === '' && beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group"
          :class="{ 'error': hardwareData.remoteKeyPort === '' && beginValidation }">
        <label for="remote-key-port">{{ $t('components.keyData.text.keyPort') }}</label>
        <input @change="hardwareDataChanged" type="number" id="remote-key-port" name="remote-key-port"
              v-model="hardwareData.remoteKeyPort" />
        <span v-if="hardwareData.remoteKeyPort === '' && beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group"
          :class="{ 'error': hardwareData.remoteKeyHost === '' && beginValidation }">
        <label for="remote-key-host">{{ $t('components.keyData.text.keyHost') }}</label>
        <input type="text" @change="hardwareDataChanged" id="remote-key-host" name="remote-key-host"
              v-model="hardwareData.remoteKeyHost" />
        <span v-if="hardwareData.remoteKeyHost === '' && beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group"
          :class="{ 'error': hardwareData.remoteKeyMask === '' && beginValidation }">
        <label for="remote-key-mask">{{ $t('components.keyData.text.keyMask') }}</label>
        <input @change="hardwareDataChanged" type="text" id="remote-key-mask" name="remote-key-mask"
              v-model="hardwareData.remoteKeyMask" />
        <span v-if="hardwareData.remoteKeyMask === '' && beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group">
        <label for="remote-ini-config">{{ $t('components.keyData.text.configurationINI') }}</label>
        <textarea rows="5" id="remote-ini-config" v-model="hardwareData.iniConfig"
                  name="remote-ini-config"></textarea>
      </div>
    </div>
    <div class="key-type-file key-type-section" v-if="deviceType === 'file'" v-cloak>
      <FileField :label="$t('components.keyData.fields.key6.label')" :sub-label="$t('components.keyData.fields.key6.subLabel')" name="key6" accept=".dat"
                :error="key6Error" @selected="onKey6FileSelected" @reset="onKey6FileReset" id="key6-upload" />
      <div class="rc-form-group"
          :class="{ 'error': fileData.signKeyIssuer === '' && beginValidation }">
        <label for="sign-key-issuer">{{ $t('components.keyData.text.keyAcsk') }}</label>
        <input type="text" id="sign-key-issuer" name="sign-key-issuer"
              v-model="fileData.signKeyIssuer" />
        <span v-if="fileData.signKeyIssuer === '' && beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group"
          :class="{ 'error': fileData.signKeyPWD === '' && beginValidation }">
        <label for="sign-key-pwd">{{ $t('components.keyData.text.passwordFileKey') }}</label>
        <input type="password" id="sign-key-pwd" name="sign-key-pwd"
              v-model="fileData.signKeyPWD" />
        <span v-if="fileData.signKeyPWD === '' && beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
    </div>

    <div class="rc-form-group allowed-keys-body" id="allowed-keys-body">
      <label>{{ $t('components.keyData.text.listAllowedKeys') }}</label>
      <a class="add-allowed-key-btn" href="#" @click.prevent="addAllowedKey">
        <i class="fa-solid fa-plus"></i>
        <div>{{ $t('components.keyData.actions.addKey') }}</div>
      </a>
      <div class="allowed-keys-row" v-for="(ak, index) in allowedKeys" :key="index"
          :class="{ 'error': beginValidation && (ak.serial === '' || ak.issuer === '') }">
        <input name="allowed-keys-issuer[]" v-model="ak.issuer"
              class="allowed-keys-input allowed-keys-issuer" aria-label="key issuer" :placeholder="$t('components.keyData.fields.issuerKey.placeholder')"
              type="text" />
        <input name="allowed-keys-serial[]"
              class="allowed-keys-input allowed-keys-serial" v-model="ak.serial" aria-label="key serial"
              :placeholder="$t('components.keyData.fields.serialNumberKey.placeholder')" type="text" />
        <button v-if="ak.removable" class="allowed-keys-remove-btn" type="button"
                @click="removeAllowedKey(ak)">-</button>
      </div>
    </div>
  </div>
  <Banner
    v-else
    :description="$t('components.keyData.text.pageDescriptionGlobal')"
  />
</template>
