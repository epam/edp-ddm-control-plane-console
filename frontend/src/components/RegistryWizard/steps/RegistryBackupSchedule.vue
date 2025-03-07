<script setup lang="ts">
import { onMounted, ref, toRefs } from 'vue';
import * as yup from 'yup';
import { useForm } from 'vee-validate';
import { parseCronExpression } from 'cron-schedule';

import Typography from '@/components/common/Typography.vue';
import TextField from '@/components/common/TextField.vue';
// @ts-ignore
import RegistryBackupSavePlaceModal from '@/components/RegistryBackupSavePlaceModal.vue';
import RegistryBackupDeletePlaceModal from '@/components/RegistryBackupDeletePlaceModal.vue';
import { getFormattedDatePrecise } from '@/utils';

interface RegistryEditTemplateVariables {
  templateVariables: {
    registryValues: {
      global: {
        registryBackup: {
          enabled: boolean;
          schedule: string;
          expiresInDays: string;
          obc: {
            cronExpression: string;
            backupBucket: string;
            endpoint: string;
          }
        }
      }
    },
    model: {
      OBCLogin: string;
      OBCPassword: string;
    }
  }
}

const parseCronExpressionRules = () => {
  return yup.string()
    .required()
    .test({
      message: 'parseCronExpression',
      test: function (value) {
        try {
          parseCronExpression(value || "");
          return true;
        }
        catch (e: any) {
          return false;
        }
      },
    });
};

const props = defineProps<RegistryEditTemplateVariables>();
const { templateVariables } = toRefs(props);
const registryValues = templateVariables?.value?.registryValues || {};
const model = templateVariables?.value?.model || {};
const beginValidation = ref(false);
const nextDates = ref([] as string[]);
const registryBackupNextDates = ref([] as string[]);
const backupPlacePopupShow = ref(false);
const backupDeletePlacePopupShow = ref(false);
const enabled = ref(registryValues.global?.registryBackup.enabled);
const obcBackupBucket = ref(registryValues.global?.registryBackup.obc.backupBucket);
const obcEndpoint = ref(registryValues.global?.registryBackup.obc.endpoint);
const obcLogin = ref(model.OBCLogin);
const obcPassword = ref(model.OBCPassword);

yup.addMethod(yup.string, "parseCronExpression", function (errorMessage) {
  return this.test(`parse-cron-expression`, errorMessage, function (value) {
    const { path, createError } = this;
    try {
      parseCronExpression(value || "");
      return true;
    }
    catch (e: any) {
      return createError({ path, message: errorMessage });
    }
  });
});

const validationSchema = yup.object({
  cronSchedule: parseCronExpressionRules(),
  days: yup.string().required().matches(/^[1-9]+$/),
  obcCronExpression: parseCronExpressionRules(),
});

const { useFieldModel, validate, errors } = useForm({
  validationSchema, initialValues: {
    cronSchedule: registryValues.global?.registryBackup.schedule,
    days: registryValues.global?.registryBackup.expiresInDays,
    obcCronExpression: registryValues.global?.registryBackup.obc.cronExpression,
  }
});

const [
  cronSchedule,
  days,
  obcCronExpression,
] = useFieldModel([
  'cronSchedule',
  'days',
  'obcCronExpression',
]);

function validator() {
  beginValidation.value = true;
  return new Promise((resolve) => {
    if (!enabled.value) {
      return resolve(true);
    }
    validate().then((res) => {
      if (res.valid) {
        beginValidation.value = false;
        resolve(true);
      }
    });
  });
}

defineExpose({
  validator
});

onMounted(()=> {
  cronExpressionChange();
  backupCronExpressionChange();
});

function showBackupPlaceModal() {
  backupPlacePopupShow.value = true;
}

function hideBackupPlaceModal() {
  backupPlacePopupShow.value = false;
}

function showBackupDeletePlaceModal() {
  backupDeletePlacePopupShow.value = true;
}

function hideBackupDeletePlaceModal() {
  backupDeletePlacePopupShow.value = false;
}

function backupPlaceSubmit(data: { backupBucket: string; endpoint: string; login: string; password: string; }) {
  obcBackupBucket.value = data.backupBucket;
  obcEndpoint.value = data.endpoint;
  obcLogin.value = data.login;
  obcPassword.value = data.password;
  backupPlacePopupShow.value = false;
}

function backupDeletePlaceSubmit() {
  obcBackupBucket.value = '';
  obcEndpoint.value = '';
  obcLogin.value = '';
  obcPassword.value = '';
  
  backupDeletePlacePopupShow.value = false;
}

function enabledChange () {
  beginValidation.value = false;
  cronSchedule.value = registryValues.global?.registryBackup.schedule;
  days.value = registryValues.global?.registryBackup.expiresInDays || "";
  obcCronExpression.value = registryValues.global?.registryBackup.obc.cronExpression;
  obcBackupBucket.value = registryValues.global?.registryBackup.obc.backupBucket;
  obcEndpoint.value = registryValues.global?.registryBackup.obc.endpoint;
  obcLogin.value = model.OBCLogin;
  obcPassword.value = model.OBCPassword;
}

function cronExpressionChange () {
  nextDates.value = [];
  try {
    const cron = parseCronExpression(cronSchedule.value);
    let dt = new Date();
    for (let i = 0; i < 3; i++) {
      const next = cron?.getNextDate(dt);
      nextDates.value.push(getFormattedDatePrecise(next?.toUTCString()));
      dt = next;
    }
  }
  catch (e: any) {
    nextDates.value = [];
  }
}

function backupCronExpressionChange () {
  registryBackupNextDates.value = [];
  try {
    const cron = parseCronExpression(obcCronExpression.value);
    let dt = new Date();
    for (let i = 0; i < 3; i++) {
      const next = cron?.getNextDate(dt);
      registryBackupNextDates.value.push(getFormattedDatePrecise(next?.toUTCString()));
      dt = next;
    }
  }
  catch (e: any) {
    registryBackupNextDates.value = [];
  }
}
</script>

<template>
  <div class="form-group">
    <Typography variant="h3">{{ $t('components.registryBackupSchedule.title') }}</Typography>
  </div>
  <Typography variant="bodyText">{{ $t('components.registryBackupSchedule.text.abilitySpecifySchedule') }}</Typography>
  <div class="toggle-switch backup-switch">
      <input v-model="enabled" :onChange="enabledChange" class="switch-input"
              type="checkbox" id="backup-schedule-switch-input"  name="backup-schedule-enabled" />
      <label for="backup-schedule-switch-input">Toggle</label>
      <span>{{ $t('components.registryBackupSchedule.text.configureBackup') }}</span>
  </div>

  <div v-show="enabled">
    <div class="form-group">
      <Typography variant="h5" upperCase>{{ $t('components.registryBackupSchedule.text.backupTheRegistry') }}</Typography>
    </div>
    <div class="form-group">
      <TextField
        :label="$t('components.registryBackupSchedule.fields.schedule.label')"
        name="cron-schedule"
        placeholder="5 4 * * *"
        :description="$t('components.registryBackupSchedule.fields.schedule.description')"
        v-model="cronSchedule"
        :error="beginValidation ? errors.cronSchedule : ''"
        required
        @change="cronExpressionChange"
      />
    </div>
    <div v-show="nextDates.length" class="form-group">
        <label>{{ $t('components.registryBackupSchedule.text.nextBackupStarts') }}</label>
        <ul class="cron-next-dates">
            <li v-for="date in nextDates" v-bind:key="date">
              <Typography variant="bodyText">{{ date }}</Typography>
            </li>
        </ul>
    </div>

    <div class="form-group">
      <TextField
        :label="$t('components.registryBackupSchedule.fields.storageDays.label')"
        name="cron-schedule-days"
        placeholder="3"
        :description="$t('components.registryBackupSchedule.fields.storageDays.description')"
        v-model="days"
        :error="beginValidation ? errors.days : ''"
        required
      />
    </div>

    <div class="form-group">
      <Typography variant="h5" upperCase>{{ $t('components.registryBackupSchedule.text.backupObjectReplications') }}</Typography>
    </div>
    <div class="form-group">
      <TextField
        :label="$t('components.registryBackupSchedule.fields.cronExpression.label')"
        name="registry-backup-obc-cron-expression"
        placeholder="30 17 * * *"
        :description="$t('components.registryBackupSchedule.fields.cronExpression.description')"
        v-model="obcCronExpression"
        :error="beginValidation ? errors.obcCronExpression : ''"
        @change="backupCronExpressionChange"
      />
      <Typography variant="small">30 17 * * *</Typography>
    </div>
    <div v-show="registryBackupNextDates.length" class="form-group">
      <label>{{ $t('components.registryBackupSchedule.text.replicationBackupRun') }}</label>
      <ul class="cron-next-dates">
          <li v-for="date in registryBackupNextDates" v-bind:key="date">
            <Typography variant="bodyText">{{ date }}</Typography>
          </li>
      </ul>
    </div>

    <div>
      <Typography variant="subheading">{{ $t('components.registryBackupSchedule.text.storageLocationBackup') }}</Typography>
      <div class="rc-form-backup-obc" v-if="obcEndpoint">
        <div>
          <div class="bucket-field">
            <Typography variant="small">{{ $t('components.registryBackupSchedule.text.bucketName') }}</Typography>
            <Typography variant="bodyText">{{ obcBackupBucket }}</Typography>
          </div>
          <div class="endpoint-field">
            <Typography variant="small">Endpoint</Typography>
            <Typography variant="bodyText">{{ obcEndpoint }}</Typography>
          </div>
          <input type="hidden" name="registry-backup-obc-backup-bucket" v-model="obcBackupBucket" disbaled />
          <input type="hidden" name="registry-backup-obc-endpoint" v-model="obcEndpoint" disbaled />
          <input type="hidden" name="registry-backup-obc-login" v-model="obcLogin" disbaled />
          <input type="hidden" name="registry-backup-obc-password" v-model="obcPassword" disbaled />
        </div>

        <div class="buttom-group">
          <a href="#" @click.stop.prevent="showBackupPlaceModal"
            class="icon-button">
            <img alt="edit button" src="@/assets/img/action-edit.png" />
          </a>
          <a href="#" @click.stop.prevent="showBackupDeletePlaceModal"
            class="icon-button">
            <img alt="delete button" src="@/assets/img/action-delete.png" />
          </a>
        </div>
      </div>
      <div class="rc-form-backup-obc-empty" v-if="!obcEndpoint">
        <Typography variant="bodyText">{{ $t('components.registryBackupSchedule.text.defaultValuesSpecifiedDeployment') }}</Typography>
        <div>
          <a href="#" @click.stop.prevent="showBackupPlaceModal" class="icon-button set-data-button">
            <img alt="edit button" src="@/assets/img/action-edit.png" />
            <Typography variant="buttonText" upperCase>{{ $t('components.registryBackupSchedule.actions.setOwnValues') }}</Typography>
          </a>
        </div>
      </div>
    </div>
  </div>

  <RegistryBackupSavePlaceModal
    :backupPlacePopupShow="backupPlacePopupShow"
    :initialData="{
      backupBucket: obcBackupBucket, endpoint: obcEndpoint, login: obcLogin, password: obcPassword,
    }"
    @submit-data="backupPlaceSubmit"
    @hide-backup-place-modal="hideBackupPlaceModal"
  />

  <RegistryBackupDeletePlaceModal
    :backupDeletePlacePopupShow="backupDeletePlacePopupShow"
    @submit-data="backupDeletePlaceSubmit"
    @hide-backup-delete-place-modal="hideBackupDeletePlaceModal"
  />
</template>

<style lang="scss" scoped>

.endpoint-field {
  margin-top: 16px;
}

.form-group {
  margin-bottom: 24px;

  label {
    font-size: 16px;
    font-weight: bold;
    margin: 0 0 8px 0;
  }
}

.rc-form-backup-obc-empty {
  width: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  margin-top: 10px;
  padding: 8px 0px;
  gap: 8px;
  border-width: 1px 0px;
  border-style: solid;
  border-color: $grey-border-color;
}

.rc-form-backup-obc {
  width: 100%;
  box-sizing: border-box;
  display: flex;
  justify-content: space-between;
  padding: 8px 0px;
  margin-top: 8px;
  gap: 8px;
  border-width: 1px 0px;
  border-style: solid;
  border-color: $grey-border-color;

  .buttom-group {
    display: flex;
    margin-top: 10px;
  }

  .rc-form-backup-obc-data {
    font-weight: 300;
    font-size: 16px;
    line-height: 24px;
  }
}

.set-data-button {
  margin: 10px;
}

.icon-button {
  display: flex;
  align-items: baseline;
  margin-left: 20px;
}

.icon-button:hover {
  text-decoration: none;
}

.icon-button > img {
  height: 18px;
  margin-right: 13px;
}

</style>
