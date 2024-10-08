<script setup lang="ts">
import { toRefs } from 'vue';

interface MRTableProps {
  mergeRequests: any;
  inPlatform?: boolean;
  mrAvailable?: string;
}

const props = defineProps<MRTableProps>();
const { inPlatform, mergeRequests, mrAvailable } = toRefs(props);
</script>

<script lang="ts">
import $ from 'jquery';
import 'datatables.net-dt';
import {
  getDateTimestamp,
  getFormattedDate,
  getMergeRequestAction,
  getMergeRequestPlatformAction,
  getMergeRequestName,
  getMergeRequestStatus,
  mrIsInProgress,
} from '@/utils';

export default {
  methods: {
    onViewClick(url: string) {
      this.$emit('onViewClick', url);
    },
    mrRefreshClick() {
      window.localStorage.setItem("mr-scroll", "true");
    },
  },
  mounted() {
    $("#mr-table").DataTable({
        ordering: true,
        paging: true,
        columnDefs: [
            { orderable: false, targets: 4 },
        ],
        order: [[0, 'desc']],
        language: {
            "processing": "Зачекайте...",
            "lengthMenu": "Показати _MENU_ записів",
            "zeroRecords": "Записи відсутні.",
            "info": "Записи з _START_ по _END_ із _TOTAL_ записів",
            "infoEmpty": "Записи з 0 по 0 із 0 записів",
            "infoFiltered": "(відфільтровано з _MAX_ записів)",
            "search": "Пошук:",
            "paginate": {
                "first": "Перша",
                "previous": "Попередня",
                "next": "Наступна",
                "last": "Остання"
            },
            "aria": {
                "sortAscending": ": активувати для сортування стовпців за зростанням",
                "sortDescending": ": активувати для сортування стовпців за спаданням"
            }
        }
    });
  }
};
</script>

<style scoped>
.mr-refresh {
  margin-left: 5px;
}
</style>

<template>
  <table class="rg-info-table rg-info-table-config" id="mr-table">
      <thead>
          <tr>
              <th>Дата</th>
              <th>Запит</th>
              <th>Операція</th>
              <th>Статус</th>
              <th></th>
          </tr>
      </thead>
      <tbody>
          <tr v-for="($al, $index) in mergeRequests" :key="$index">
              <td :data-order="getDateTimestamp($al.metadata.creationTimestamp)">{{getFormattedDate($al.metadata.creationTimestamp)}}</td>
              <td>{{ inPlatform ? $al.metadata.name : getMergeRequestName($al) }}</td>
              <td>{{ inPlatform ? getMergeRequestPlatformAction($al) : getMergeRequestAction($al) }}</td>
              <td class="mr-status">
                {{ getMergeRequestStatus($al) }}
                <a title="Оновити" class="mr-refresh" href="" @click="mrRefreshClick" v-if="mrIsInProgress($al)"><i class="fa-solid fa-arrow-rotate-right"></i></a>
              </td>
              <td class="mr-actions">
                  <i v-if="!inPlatform && !mrAvailable && $al.status.value === 'NEW'" title="Реєстр в процесі оновлення" class="fa-solid fa-lock"></i>

                  <span v-if="$al.status.changeUrl && (inPlatform || mrAvailable || $al.status.value !== 'NEW')">
                      <a title="Переглянути"
                          @click.stop.prevent="onViewClick(`/admin/change/${$al.status.changeId}`)"
                          :href="`/admin/change/${$al.status.changeId}`">
                          <i class="fa-solid fa-eye fa-lg"></i>
                      </a>

                      <a :href="$al.status.changeUrl" target="_blank">
                          <img style="vertical-align: sub;" title="Переглянути в Gerrit" alt="vcs"
                              src="@/assets/img/action-link.png" />
                      </a>
                  </span>
              </td>
          </tr>
      </tbody>
  </table>
</template>
