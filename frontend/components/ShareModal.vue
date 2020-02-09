<template>
  <b-modal ref="share-modal" @ok="share" size="xl" title="Share">
    <b-container>
      <b-container v-if="newLinkShareAccess !== shareAccessLevels[0]">
        <b-dropdown
          v-model="newLinkShareAccess"
          @change="changedLinkShareAccess"
          :text="`Anyone with the link can ${newLinkShareAccess}`"
        >
          <b-dropdown-item>View</b-dropdown-item>
          <b-dropdown-item>Edit</b-dropdown-item>
        </b-dropdown>
        <b-form-input
          id="shareableLink"
          v-model="shareableLink"
          :disabled="true"
          type="text"
        />
        <b-btn @click="copyToClipboard" class="mb-3" variant="primary">
          Copy Shareable Link
        </b-btn>
      </b-container>
      <b-btn v-else @click="enableLinkShare">
        Enable Link Sharing
      </b-btn>
      <b-row>
        <client-only>
          <multiselect
            v-model="newAccess"
            :options="newAccessOptions"
            :multiple="true"
            :taggable="true"
            @tag="newUserAccess"
            track-by="email"
            label="email"
          />
        </client-only>
        <b-dropdown v-model="newAccessLevel" :text="newAccessLevel">
          <b-dropdown-item>View</b-dropdown-item>
          <b-dropdown-item>Edit</b-dropdown-item>
        </b-dropdown>
      </b-row>
      <b-btn v-if="!editAccessLevels" @click="editAccess">
        Edit Access
      </b-btn>
      <b-container v-else>
        <b-card
          v-for="(access, index) in changedAccess"
          :key="`access-${index}`"
        >
          {{ access.email }}
          <b-dropdown
            v-model="changedAccess[index].type"
            @change="(evt) => changedUserAccess(evt, index)"
            :text="changedAccess[index].type"
          >
            <b-dropdown-item>View</b-dropdown-item>
            <b-dropdown-item>Edit</b-dropdown-item>
          </b-dropdown>
        </b-card>
      </b-container>
    </b-container>
  </b-modal>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import Multiselect from 'vue-multiselect'
import * as clipboard from 'clipboard-polyfill'
import { validTypes } from '~/assets/config'
import { clone } from '~/assets/utils'
const shareAccessLevels = ['none', 'view', 'edit']
const shortLinkURL = process.env.shortlinkurl
// TODO - check if user exists before sharing with them
export default Vue.extend({
  name: 'Share',
  components: {
    Multiselect
  },
  props: {
    type: {
      type: String,
      default: null,
      validator: (value) => validTypes.includes(value)
    },
    id: {
      type: String,
      default: null,
      validator: (value) => value && value.length > 0
    }
  },
  data() {
    return {
      updates: { // updates
        access: [],
        public: null,
        linkShareAccess: null
      },
      newAccessOptions: [], // options for new access
      currentAccess: [], // this is the original access
      changedAccess: [], // this is current access v-model
      newAccess: [], // this is added access
      newAccessLevel: shareAccessLevels[2],
      owner: '',
      public: '',
      shareableLink: '',
      currentLinkShareAccess: '',
      newLinkShareAccess: '',
      editAccessLevels: false,
      shareAccessLevels
    }
  },
  mounted() {
    const setData = (data) => {
      this.currentAccess = data.access
      this.changedAccess = clone(this.currentAccess)
      this.owner = data.owner
      this.currentLinkShareAccess = data.linkaccess.type
      this.newLinkShareAccess = this.currentLinkShareAccess
      this.shareableLink = `${shortLinkURL}/${data.linkaccess.shortlink}`
      this.public = data.public
    }
    if (this.type === validTypes[0]) {
      // project type
      this.$apollo
        .query({
          query: gql`
            query project {
              project {
                owner
                public
                linkaccess {
                  shortlink
                  type
                }
                access {
                  id
                  type
                }
              }
            }
          `,
          variables: {},
          fetchPolicy: 'network-only'
        })
        .then(({ data }) => {
          if (data.project) {
            setData(data.project)
          } else {
            this.$bvToast.toast('cannot find access info', {
              variant: 'danger',
              title: 'Error'
            })
          }
        })
        .catch((err) => {
          this.$bvToast.toast(err, {
            variant: 'danger',
            title: 'Error'
          })
        })
    } else {
      // form type
      this.$apollo
        .query({
          query: gql`
            query form {
              form {
                owner
                public
                linkaccess {
                  shortlink
                  type
                }
                access {
                  id
                  type
                }
              }
            }
          `,
          variables: {},
          fetchPolicy: 'network-only'
        })
        .then(({ data }) => {
          if (data.form) {
            setData(data.form)
          } else {
            this.$bvToast.toast('cannot find access info', {
              variant: 'danger',
              title: 'Error'
            })
          }
        })
        .catch((err) => {
          this.$bvToast.toast(err, {
            variant: 'danger',
            title: 'Error'
          })
        })
    }
  },
  methods: {
    show() {
      if (this.$refs['share-modal']) {
        this.$refs['share-modal'].show()
      } else {
        this.$bvToast.toast('cannot find share modal', {
          variant: 'danger',
          title: 'Error'
        })
      }
    },
    share(evt) {
      evt.preventDefault()
      for (let i = 0; i < this.newAccess.length; i++) {
        if (this.updates.access.findIndex((elem) => elem.id === this.newAccess[i].id) >= 0) {
          this.updates.access[i].type = this.newAccessLevel
        } else {
          this.updates.access.push({
            id: this.newAccess[i].id,
            type: this.newAccessLevel
          })
        }
      }
      const finished = () => {
        this.$nextTick(() => {
          if (this.$refs['share-modal']) {
            this.$refs['share-modal'].hide()
          }
        })
      }
      if (this.updates.public || this.updates.linkShareAccess || this.updates.access.length > 0) {
        if (this.type === validTypes[0]) {
          // project type
          this.$apollo.mutate({mutation: gql`
            mutation updateProject($access: AccessInputType, $public: String, $linkaccess: String) {
              updateProject(access: $access, public: $public, linkaccess: $linkaccess) {
                id
              }
            }
            `, variables: {
              access: this.updates.access,
              public: this.updates.public,
              linkaccess: this.updates.linkShareAccess
            }})
            .then(({ data }) => {
              finished()
            }).catch(err => {
              this.$bvToast.toast(`found error: ${err.message}`, {
                variant: 'danger',
                title: 'Error'
              })
            })
        } else {
          this.$apollo.mutate({mutation: gql`
            mutation updateForm($access: AccessInputType, $public: String, $linkaccess: String) {
              updateForm(access: $access, public: $public, linkaccess: $linkaccess) {
                id
              }
            }
            `, variables: {
              access: this.updates.access,
              public: this.updates.public,
              linkaccess: this.updates.linkShareAccess
            }})
            .then(({ data }) => {
              finished()
            }).catch(err => {
              this.$bvToast.toast(`found error: ${err.message}`, {
                variant: 'danger',
                title: 'Error'
              })
            })
        }
      } else {
        finished()
      }
    },
    changedLinkShareAccess(evt) {
      if (this.newLinkShareAccess !== this.currentLinkShareAccess) {
        this.updates.linkShareAccess = this.newLinkShareAccess
      } else {
        this.updates.linkShareAccess = null
      }
    },
    newUserAccess(email) {
      // https://stackoverflow.com/a/55636364/8623391
      // TODO - get id based on email, validate email
      const tag = {
        email
      }
      this.newAccessOptions.push(tag)
      this.newAccess.push(tag)
    },
    changedUserAccess(evt, index) {
      const currentAccessElem = this.currentAccess.find(elem => elem.id === this.changedAccess[index].id)
      const changed = this.changedAccess[index].type !== currentAccessElem.type
      const accessIndex = this.updates.access.findIndex(elem => elem.id === this.changedAccess[index].id)
      if (accessIndex >= 0) {
        if (changed) {
          this.updates.access[accessIndex].type = this.changedAccess[index].type
        } else {
          this.updates.splice(accessIndex, 1)
        }
      } else if (changed) {
        this.updates.access.push({
          id: currentAccessElem.id,
          type: this.changedAccess[index].type
        })
      }
    },
    enableLinkShare(evt) {
      evt.preventDefault()
      this.newLinkShareAccess = this.shareAccessLevels[1]
    },
    copyToClipboard(evt) {
      evt.preventDefault()
      const dt = new clipboard.DT()
      dt.setData('text', this.shareableLink)
      clipboard.write(dt)
      this.$bvToast.toast('Link copied!', {
        variant: 'success',
        title: 'Success'
      })
    },
    editAccess(evt) {
      evt.preventDefault()
      this.editAccessLevels = true
    }
  }
})
</script>

<style lang="scss"></style>
