<script setup lang="ts">
import axios, { AxiosError } from 'axios';
import TextField from '@/components/common/TextField.vue';
import * as yup from 'yup';
import { useForm } from 'vee-validate';
import { parseCronExpression } from 'cron-schedule';
import { ref, toRefs } from 'vue';

interface Data {
  NexusSchedule: string;
  NexusExpiresInDays: string;
  ControlPlaneSchedule: string;
  ControlPlaneExpiresInDays: string;
  UserManagementSchedule: string;
  UserManagementExpiresInDays: string;
  MonitoringSchedule: string;
  MonitoringExpiresInDays: string;
}

interface BackupBlockProps {
  backupSchedule: Data;
}

const props = defineProps<BackupBlockProps>();
const { backupSchedule } = toRefs(props);
const beginValidation = ref(false);

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

const validationSchema = yup.object({
  NexusSchedule: parseCronExpressionRules(),
  NexusExpiresInDays: yup.string().required().matches(/^[1-9]+$/),
  ControlPlaneSchedule: parseCronExpressionRules(),
  ControlPlaneExpiresInDays: yup.string().required().matches(/^[1-9]+$/),
  UserManagementSchedule: parseCronExpressionRules(),
  UserManagementExpiresInDays: yup.string().required().matches(/^[1-9]+$/),
  MonitoringSchedule: parseCronExpressionRules(),
  MonitoringExpiresInDays: yup.string().required().matches(/^[1-9]+$/),
});

const { handleSubmit, useFieldModel, setErrors, validate, values, errors } = useForm({
  validationSchema, initialValues: backupSchedule, validateOnMount: false,
});

const [
  NexusSchedule,
  NexusExpiresInDays,
  ControlPlaneSchedule,
  ControlPlaneExpiresInDays,
  UserManagementSchedule,
  UserManagementExpiresInDays,
  MonitoringSchedule,
  MonitoringExpiresInDays,
] = useFieldModel([
  'NexusSchedule',
  'NexusExpiresInDays',
  'ControlPlaneSchedule',
  'ControlPlaneExpiresInDays',
  'UserManagementSchedule',
  'UserManagementExpiresInDays',
  'MonitoringSchedule',
  'MonitoringExpiresInDays',
]);

NexusExpiresInDays.value = NexusExpiresInDays.value === '0' ? '' : NexusExpiresInDays.value;
ControlPlaneExpiresInDays.value = ControlPlaneExpiresInDays.value === '0' ? '' : ControlPlaneExpiresInDays.value;
UserManagementExpiresInDays.value = UserManagementExpiresInDays.value === '0' ? '' : UserManagementExpiresInDays.value;
MonitoringExpiresInDays.value = MonitoringExpiresInDays.value === '0' ? '' : MonitoringExpiresInDays.value;

const submit = handleSubmit(() => {
  let formData = new FormData();

  formData.append("nexus-schedule", values.NexusSchedule);
  formData.append("nexus-expires-in-days", values.NexusExpiresInDays);
  formData.append("control-plane-schedule", values.ControlPlaneSchedule);
  formData.append("control-plane-expires-in-days", values.ControlPlaneExpiresInDays);
  formData.append("control-plane-expires-in-days", values.ControlPlaneExpiresInDays);
  formData.append("user-management-schedule", values.UserManagementSchedule);
  formData.append("user-management-expires-in-days", values.UserManagementExpiresInDays);
  formData.append("monitoring-schedule", values.MonitoringSchedule);
  formData.append("monitoring-expires-in-days", values.MonitoringExpiresInDays);

  
  axios.post('/admin/cluster/backup-schedule', formData, {
      headers: {
          'Content-Type': 'multipart/form-data'
      }
  }).then(() => {
    beginValidation.value = false;
    window.location.assign('/admin/cluster/management');
  }).catch(({ response }: AxiosError<any>) => {
    setErrors(response?.data.errors);
  });
});

function onSubmit() {
  beginValidation.value = true;
  submit();
}

</script>

<template>
  <h2>{{ $t('domains.cluster.backup.title') }}</h2>
  <form @submit.prevent="onSubmit" id="backup-schedule-form" class="registry-create-form wizard-form">
    <h3>Nexus</h3>
    <div class="form-group">
      <TextField
        :label="$t('domains.cluster.backup.fields.nexusSchedule.label')"
        name="nexus-schedule"
        placeholder="0 10 * * *"
        :description="$t('domains.cluster.backup.fields.nexusSchedule.description')"
        v-model="NexusSchedule"
        :error="beginValidation ? errors.NexusSchedule : ''"
        required
        @change="validate"
      />
    </div>
    <div class="form-group">
      <TextField
        :label="$t('domains.cluster.backup.fields.nexusTime.label')"
        name="nexus-expires-in-days"
        placeholder="5"
        :description="$t('domains.cluster.backup.fields.nexusTime.description')"
        v-model="NexusExpiresInDays"
        :error="beginValidation ? errors.NexusExpiresInDays : ''"
        required
        @change="validate"
      />
    </div>

    <h3>Control Plane</h3>
    <div class="form-group">
      <TextField
        :label="$t('domains.cluster.backup.fields.controlPlaneSchedule.label')"
        name="control-plane-schedule"
        placeholder="0 10 * * *"
        :description="$t('domains.cluster.backup.fields.controlPlaneSchedule.description')"
        v-model="ControlPlaneSchedule"
        :error="beginValidation ? errors.ControlPlaneSchedule : ''"
        required
        @change="validate"
      />
    </div>
    <div class="form-group">
      <TextField
        :label="$t('domains.cluster.backup.fields.controlPlaneTime.label')"
        name="control-plane-expires-in-days"
        placeholder="5"
        :description="$t('domains.cluster.backup.fields.controlPlaneTime.description')"
        v-model="ControlPlaneExpiresInDays"
        :error="beginValidation ? errors.ControlPlaneExpiresInDays : ''"
        required
        @change="validate"
      />
    </div>

    <h3>User Management</h3>
    <div class="form-group">
      <TextField
        :label="$t('domains.cluster.backup.fields.managementSchedule.label')"
        name="user-management-schedule"
        placeholder="0 10 * * *"
        :description="$t('domains.cluster.backup.fields.managementSchedule.description')"
        v-model="UserManagementSchedule"
        :error="beginValidation ? errors.UserManagementSchedule : ''"
        required
        @change="validate"
      />
    </div>
    <div class="form-group">
      <TextField
        :label="$t('domains.cluster.backup.fields.managementTime.label')"
        name="user-management-expires-in-days"
        placeholder="5"
        :description="$t('domains.cluster.backup.fields.managementTime.description')"
        v-model="UserManagementExpiresInDays"
        :error="beginValidation ? errors.UserManagementExpiresInDays : ''"
        required
        @change="validate"
      />
    </div>

    <h3>Monitoring</h3>
    <div class="form-group">
      <TextField
        :label="$t('domains.cluster.backup.fields.monitoringSchedule.label')"
        name="monitoring-schedule"
        placeholder="0 10 * * *"
        :description="$t('domains.cluster.backup.fields.monitoringSchedule.description')"
        v-model="MonitoringSchedule"
        :error="beginValidation ? errors.MonitoringSchedule : ''"
        required
        @change="validate"
      />
    </div>
    <div class="form-group">
      <TextField
        :label="$t('domains.cluster.backup.fields.monitoringTime.label')"
        name="monitoring-expires-in-days"
        placeholder="5"
        :description="$t('domains.cluster.backup.fields.monitoringTime.description')"
        v-model="MonitoringExpiresInDays"
        :error="beginValidation ? errors.MonitoringExpiresInDays : ''"
        required
        @change="validate"
      />
    </div>

    <div class="rc-form-group">
      <button type="submit" name="submit" onclick="window.localStorage.setItem('mr-scroll', 'true');">
        {{ $t('actions.confirm') }}
      </button>
    </div>
  </form>
</template>


<style lang="scss" scoped>

.form-group {
  margin-bottom: 24px;
}

</style>

