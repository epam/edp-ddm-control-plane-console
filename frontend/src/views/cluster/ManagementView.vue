<script setup lang="ts">
import { inject } from 'vue';
import i18n, { getNamespaceMessage } from '@/localization';

interface ManagementTemplateVariables {
    canUpdateCluster: any;
    codebase: any;
    version: any;
    admins: any;
    cidr: any;
    branches: any;
    mergeRequests: any;
    edpComponents: any;
    jenkinsURL: any;
    gerritURL: any;
    platformCentralComponents: any;
    platformAdministrationComponents: any;
    platformOperationalComponents: any;
    hasUpdate: any;
    language: keyof typeof LANGUAGES;
}

const variables = inject('TEMPLATE_VARIABLES') as ManagementTemplateVariables;
const canUpdateCluster = variables?.canUpdateCluster;
const codebase = variables?.codebase;
const version = variables?.version;
const admins = variables?.admins;
const cidr = variables?.cidr;
const hasUpdate = variables?.hasUpdate;
const branches = variables?.branches;
const mergeRequests = variables?.mergeRequests;
const edpComponents = variables?.edpComponents;
const jenkinsURL = variables?.jenkinsURL;
const gerritURL = variables?.gerritURL;
const platformCentralComponents = variables?.platformCentralComponents;
const platformAdministrationComponents = variables?.platformAdministrationComponents;
const platformOperationalComponents = variables?.platformOperationalComponents;
const language = variables?.language || i18n.global.locale;

</script>
<script lang="ts">
import { defineComponent } from 'vue';
import $ from 'jquery';

import { LANGUAGES } from '@/constants/cluster';
import { getFormattedDate, getGerritURL, getImageUrl, getJenkinsURL, getStatusTitle } from '@/utils';
import MergeRequestsTable from '@/components/MergeRequestsTable.vue';

export default defineComponent({
    data() {
        return {
            backdropShow: false,
            mrView: false,
            mrSrc: '',
            activeTab: 'info',
            accordion: {
              general: true,
              configuration: false,
              mergeRequests: false,
            },
        };
    },
    methods: {
        hasNewMergeRequests() {
          let statuses = $(".mr-status");
          for (let i = 0; i < statuses.length; i++) {
            const statusHtml = $(statuses[i]).html().trim();
            if (statusHtml === this.$t('pages.clusterManagement.text.new') || statusHtml.indexOf('mr-refresh') !== -1) {
              return true;
            }
          }

          return false;
        },
        showMrView(src: string) {
            this.mrView = true;
            this.backdropShow = true;
            $("body").css("overflow", "hidden");
            window.scrollTo(0, 0);
            this.mrSrc = src;
        },
        hideMrView() {
            $("body").css("overflow", "scroll");
            this.backdropShow = false;
            this.mrView = false;
            let mrFrame = this.$refs.mrIframe;
            if ((mrFrame as any).src !== (mrFrame as any).contentWindow.location.href) {
                document.location.reload();
            }
        },
        selectTab(tabName: string) {
          this.activeTab = tabName;
        },
        isActiveTab(tabName: string) {
          return this.activeTab === tabName;
        },
    },
    components: { MergeRequestsTable },
    mounted() {
        const scroll = window.localStorage.getItem("mr-scroll");
        if (scroll) {
          this.accordion.mergeRequests = true;
          window.localStorage.removeItem("mr-scroll");
          this.$nextTick(() => {
            document.getElementById('merge-requests-body')?.scrollIntoView({
              behavior: "smooth", block: "end", inline: "nearest" });
          });
        }
    },
});
</script>

<style scoped>
  .rg-info-block-header:hover {
    background: #00689B;
  }
  .rg-info-block-header {
    transition: 0.5s;
  }
</style>

<template>
    <div class="registry cluster" id="registry-view">
        <div class="registry-header registry-header-view cluster">
            <h1>{{ $t('pages.clusterManagement.title') }}</h1>
          <div class="registry-view-actions">
            <template v-if="hasUpdate">
              <a :href="`/admin/cluster/edit#upgrade`"
                 class="registry-add">
                <i class="fa-solid fa-arrow-up"></i>
                <span>{{ $t('actions.refresh') }}</span>
              </a>
            </template>
            <a v-if="canUpdateCluster" href="/admin/cluster/edit" class="registry-add">
                <img alt="add registry" src="@/assets/img/action-edit.png" />
                <span>{{ $t('actions.edit') }}</span>
            </a>
          </div>
        </div>
        <div class="tabs">
          <div class="tab" @click="selectTab('info')" :class="{ active: isActiveTab('info') }">
            {{ $t('pages.clusterManagement.tabs.platformInfo') }}
          </div>
          <div class="tab" @click="selectTab('links')" :class="{ active: isActiveTab('links') }">
            {{ $t('pages.clusterManagement.tabs.fastLinks') }}
          </div>
        </div>

        <div class="box" v-show="isActiveTab('info')">
          <div class="rg-info-block">
              <div class="rg-info-block-header" @click="accordion.general = !accordion.general">
                  <span>{{ $t('pages.clusterManagement.text.commonInfo') }}</span>
                  <img v-if="accordion.general" src="@/assets/img/action-toggle.png" alt="toggle block" />
                  <img v-if="!accordion.general" src="@/assets/img/down.png" alt="toggle block" />
              </div>
              <div class="rg-info-block-body" v-show="accordion.general">
                  <div class="rg-info-line-horizontal">
                      <span>{{ $t('pages.clusterManagement.text.name') }}</span>
                      <span>{{ codebase.metadata.name }}</span>
                  </div>
                  <div class="rg-info-line-horizontal">
                      <span>{{ $t('pages.clusterManagement.text.version') }}</span>
                      <span>{{ version }}</span>
                  </div>
                  <div class="rg-info-line-horizontal">
                      <span>{{ $t('pages.clusterManagement.text.description') }}</span>
                      <span>{{ getNamespaceMessage('pages.clusterManagement.text', codebase.spec.description) }}</span>
                  </div>
                  <div v-if="admins" class="rg-info-line-horizontal">
                      <span>{{ $t('pages.clusterManagement.text.administrators') }}</span>
                      <span>{{ admins }}</span>
                  </div>
                  <div v-if="cidr" class="rg-info-line-horizontal">
                      <span>{{ $t('pages.clusterManagement.text.cidrValues') }}</span>
                      <span class="cidr-values">
                          <div v-for="(value, index) in cidr" :key="index" class="view-cidr">{{ value }}</div>
                      </span>
                  </div>

                  <div class="rg-info-line-horizontal">
                      <span>{{ $t('pages.clusterManagement.text.creationTimestamp') }}</span>
                      <span>{{ getFormattedDate(codebase.metadata.creationTimestamp) }}</span>
                  </div>
                  <div class="rg-info-line-horizontal" v-if="language">
                      <span>{{ $t('pages.clusterManagement.text.language') }}</span>
                      <span>
                        {{ LANGUAGES[language].name }}
                      </span>
                  </div>
              </div>
          </div>

          <div v-if="branches.length" class="rg-info-block">
              <div class="rg-info-block-header" @click="accordion.configuration = !accordion.configuration">
                  <span>{{ $t('pages.clusterManagement.text.config') }}</span>
                <img v-if="accordion.configuration" src="@/assets/img/action-toggle.png" alt="toggle block" />
                <img v-if="!accordion.configuration" src="@/assets/img/down.png" alt="toggle block" />

              </div>
              <div class="rg-info-block-body" v-show="accordion.configuration">
                  <table class="rg-info-table rg-info-table-config">
                      <thead>
                          <tr>
                              <th>{{ $t('pages.clusterManagement.table.status') }}</th>
                              <th>{{ $t('pages.clusterManagement.table.config') }}</th>
                              <th>VCS</th>
                              <th>CI</th>
                              <th>{{ $t('pages.clusterManagement.table.version') }}</th>
                          </tr>
                      </thead>
                      <tbody>
                          <tr v-for="$br in branches" :key="$br.Name">
                              <td>
                                  <img :title="getStatusTitle($br.status.value)" :alt="getStatusTitle($br.status.value)"
                                      :src="getImageUrl(`status-${$br.status.value}`)" />
                              </td>
                              <td>
                                  {{ $br.metadata.name }}
                              </td>
                              <td>
                                  <a :href="getGerritURL(gerritURL)" target="_blank">
                                      <img alt="vcs" src="@/assets/img/action-link.png" />
                                  </a>
                              </td>
                              <td>
                                  <a :href="getJenkinsURL(jenkinsURL, $br.spec.codebaseName, $br.spec.branchName)"
                                      target="_blank">
                                      <img alt="ci" src="@/assets/img/action-link.png" />
                                  </a>
                              </td>
                              <td>{{ $br.spec.version }}</td>
                          </tr>
                      </tbody>
                  </table>
              </div>
          </div>

          <div v-if="mergeRequests.length" class="rg-info-block">
              <div class="rg-info-block-header" @click="accordion.mergeRequests = !accordion.mergeRequests">
                  <span>{{ $t('pages.clusterManagement.text.requests') }}</span>
                <img v-if="accordion.mergeRequests" src="@/assets/img/action-toggle.png" alt="toggle block" />
                <img v-if="!accordion.mergeRequests" src="@/assets/img/down.png" alt="toggle block" />

              </div>
              <div id="merge-requests-body" class="rg-info-block-body mr-block-table" v-show="accordion.mergeRequests">
                  <MergeRequestsTable :merge-requests="mergeRequests" :in-platform="true" @onViewClick="showMrView"></MergeRequestsTable>
              </div>
          </div>
          <div v-if="edpComponents.length" class="dashboard-panel registry-dashboard-panel">
              <ul>
                  <li v-for="$ec in edpComponents" :key="$ec.spec.type">
                      <img :src="`data:image/svg+xml;base64,${$ec.spec.icon}`" :alt="`${$ec.spec.type} logo`" />
                      <div class="dashboard-item-content">
                          <a :href="$ec.spec.url">
                              {{ $ec.spec.type }}
                              <img src="@/assets/img/action-link.png" :alt="`${$ec.spec.type} link`">
                          </a>
                      </div>
                  </li>
              </ul>
          </div>
          <div class="popup-backdrop visible" v-cloak v-if="backdropShow"></div>

          <div style="width:80%;left:10%;height:80%;" class="popup-window admin-window visible" v-cloak v-if="mrView">
              <div class="popup-header">
                  <p>{{ $t('pages.clusterManagement.text.request') }}</p>
                  <a href="#" @click.stop.prevent="hideMrView" class="popup-close hide-popup">
                      <img alt="close popup window" src="@/assets/img/close.png" />
                  </a>
              </div>
              <div class="popup-body mr-frame-body" style="border-bottom: none;">
                  <iframe ref="mrIframe" id="mr-frame" :src="mrSrc" style="width:100%;"></iframe>
              </div>
          </div>
        </div>

        <div class="box" v-show="isActiveTab('links')">
          <template v-if="platformAdministrationComponents">
            <div class="rg-info-block">
              <div class="rg-info-block-header">
                <span>{{ $t('pages.clusterManagement.text.adminZonePlatform') }}</span>
              </div>
              <div class="rg-info-block-body mr-block-table">
                <div class="dashboard-panel registry-dashboard-panel">
                  <template v-for="$ec in platformAdministrationComponents" :key="$ec.Url">
                    <div class="list-item">
                      <img :src="`data:image/svg+xml;base64,${$ec.Icon}`" :alt="`${$ec.Type} logo`"
                           class="item-image" />
                      <div class="item-content">
                        <a target="_blank" :href="$ec.Url">
                          {{ getNamespaceMessage('domains.cluster.links.platformAdministration.title', $ec.Title) }}
                          <img src="@/assets/img/action-link.png" :alt="`${$ec.Type} link`">
                        </a>
                        <div class="description">{{ getNamespaceMessage('domains.cluster.links.platformAdministration.description', $ec.Description) }}</div>
                      </div>
                    </div>
                  </template>
                </div>
              </div>
            </div>
          </template>
          <template v-if="platformOperationalComponents">
            <div class="rg-info-block">
              <div class="rg-info-block-header">
                <span>{{ $t('pages.clusterManagement.text.operationZonePlatform') }}</span>
              </div>
              <div class="rg-info-block-body mr-block-table">
                <div class="dashboard-panel registry-dashboard-panel">
                  <template v-for="$ec in platformOperationalComponents" :key="$ec.Url">
                    <div class="list-item">
                      <img :src="`data:image/svg+xml;base64,${$ec.Icon}`" :alt="`${$ec.Type} logo`"
                           class="item-image" />
                      <div class="item-content">
                        <a target="_blank" :href="$ec.Url">
                          {{ getNamespaceMessage('domains.cluster.links.platformOperational.title', $ec.Title) }}
                          <img src="@/assets/img/action-link.png" :alt="`${$ec.Type} link`">
                        </a>
                        <div class="description">{{ getNamespaceMessage('domains.cluster.links.platformOperational.description', $ec.Description) }}</div>
                      </div>
                    </div>
                  </template>
                </div>
              </div>
            </div>
          </template>
          <template v-if="platformCentralComponents">
            <div class="rg-info-block">
              <div class="rg-info-block-header">
                <span>{{ $t('pages.clusterManagement.text.centerComponents') }}</span>
              </div>
              <div class="rg-info-block-body mr-block-table">
                <div class="dashboard-panel registry-dashboard-panel">
                  <template v-for="$ec in platformCentralComponents" :key="$ec.Url">
                    <div class="list-item">
                      <img :src="`data:image/svg+xml;base64,${$ec.Icon}`" :alt="`${$ec.Type} logo`"
                           class="item-image" />
                      <div class="item-content">
                        <a target="_blank" :href="$ec.Url" :class="{ disabled: $ec.Visible == 'false' }">
                          {{ getNamespaceMessage('domains.cluster.links.platformCentral.title', $ec.Title) }}
                          <span v-if="$ec.Visible == 'false'">({{ $t('pages.clusterManagement.text.turnedOff') }})</span>
                          <img v-else src="@/assets/img/action-link.png" :alt="`${$ec.Type} link`">
                        </a>
                        <div class="description">{{ getNamespaceMessage('domains.cluster.links.platformCentral.description', $ec.Description) }}</div>
                      </div>
                    </div>
                  </template>
                </div>
              </div>
            </div>
          </template>
        </div>
    </div>
</template>
