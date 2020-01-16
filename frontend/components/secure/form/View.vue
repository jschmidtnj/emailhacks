<template>
  <div id="view">
    <div v-if="!loading">
      <b-card no-body class="card-data shadow-lg pb-4">
        <b-card-body>
          <b-form @submit.prevent>
            <h3 class="mb-4">
              {{ name }}
            </h3>
            <div
              v-for="(item, index) in items"
              :key="`item-${index}`"
              :class="{ 'item-focus': focusIndex === index }"
            >
              <hr class="separate-items" />
              <span
                :id="`item-${index}-select-area`"
                v-touch:start="(evt) => focusItem(evt, index)"
              >
                <b-container v-if="item.type !== itemTypes[3]">
                  <h4>{{ item.question }}</h4>
                </b-container>
                <b-container v-else>
                  <p v-html="item.text" />
                </b-container>
                <b-input-group
                  v-if="
                    item.type !== itemTypes[3] && item.type !== itemTypes[6]
                  "
                  :id="`item-${index}-question`"
                >
                  <b-container>
                    <div
                      v-if="
                        item.type === itemTypes[0] || item.type === itemTypes[1]
                      "
                    >
                      <b-row
                        v-for="(option, optionIndex) in item.options"
                        :key="`item-${index}-option-${optionIndex}`"
                        class="mt-2 mb-2"
                        style="max-width:30rem;"
                      >
                        <b-col>
                          <b-form-radio v-if="item.type === itemTypes[0]">{{
                            item.options[optionIndex]
                          }}</b-form-radio>
                          <b-form-checkbox
                            v-else-if="item.type === itemTypes[1]"
                            >{{ item.options[optionIndex] }}</b-form-checkbox
                          >
                        </b-col>
                      </b-row>
                    </div>
                    <b-form-textarea
                      v-else-if="item.type === itemTypes[2]"
                      :id="`item-${index}-shortAnswer`"
                      class="mt-2 mb-2"
                      rows="3"
                      max-rows="8"
                      style="max-width:30rem;"
                    />
                    <div
                      v-else-if="item.type === itemTypes[4]"
                      class="mt-2 mb-2"
                    >
                      <b-form-checkbox
                        :id="`item-${index}-red-green`"
                        style="display: inline-block;"
                        name="red-green"
                        switch
                      />
                    </div>
                    <div
                      v-else-if="item.type === itemTypes[5]"
                      class="mt-2 mb-2"
                    >
                      <b-form-file
                        placeholder="Choose a file or drop it here..."
                        drop-placeholder="Drop file here..."
                        style="max-width:30rem;"
                      />
                    </div>
                    <div
                      v-else-if="item.type === itemTypes[6]"
                      class="mt-2 mb-2"
                    >
                      <b-container>
                        <b-row>
                          <b-col>
                            <a
                              v-if="items[index].file"
                              :href="getFileURL(index)"
                              :download="items[index].file.name"
                              class="mt-2 mb-2"
                              >Download</a
                            >
                          </b-col>
                        </b-row>
                      </b-container>
                    </div>
                  </b-container>
                </b-input-group>
              </span>
            </div>
          </b-form>
        </b-card-body>
      </b-card>
      <b-container
        v-if="!preview"
        style="margin-top: 3rem; margin-bottom: 2rem;"
      >
        <b-row>
          <b-col class="text-right">
            <b-button
              @click="submit"
              pill
              variant="primary"
              class="submit-button shadow-lg"
            >
              <client-only>
                <font-awesome-icon size="3x" icon="paper-plane" />
              </client-only>
            </b-button>
          </b-col>
        </b-row>
      </b-container>
    </div>
    <page-loading v-else :loading="true" />
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import PageLoading from '~/components/PageLoading.vue'
import { noneAccessType, defaultItem } from '~/assets/config'
import { clone, arrayMove, checkDefined } from '~/assets/utils'
const itemTypes = ['radio', 'checkbox', 'short', 'text',
  'redgreen', 'fileupload', 'fileattachment', 'media']
const defaultFile = {
  id: '',
  name: '',
  type: '',
  src: null,
  uploaded: false
}
export default Vue.extend({
  name: 'ViewForm',
  components: {
    PageLoading
  },
  props: {
    formId: {
      type: String,
      default: null
    },
    projectId: {
      type: String,
      default: null
    },
    preview: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      name: '',
      items: [],
      multiple: false,
      focusIndex: 0,
      itemTypes,
      loading: true,
      isPublic: false,
      updatesAccessToken: null,
      connectionId: null
    }
  },
  mounted() {
    if (this.formId) {
      this.$apollo.query({query: gql`
        query form($id: String!){
          form(id: $id, editAccessToken: false) {
            name
            items {
              question
              type
              options
              text
              required
              files
            }
            multiple
            files {
              id
              name
              type
              originalSrc
            }
            updatesAccessToken
          }
        }
        `, variables: {id: this.formId}})
        .then(({ data }) => {
          console.log(data.form)
          this.name = data.form.name
          this.isPublic = data.form.public === noneAccessType
          this.$store.commit('auth/setRedirectLogin', this.isPublic)
          for (let i = 0; i < data.form.items.length; i++) {
            if (data.form.items[i].files.length === 0) {
              data.form.items[i].files = [clone(defaultFile)]
            } else {
              const newFiles = []
              for (let j = 0; j < data.form.items[i].files.length; j++) {
                const fileData = data.form.files[data.form.items[i].files[j]]
                const fileObj = clone(defaultFile)
                for (const key in fileData) {
                  fileObj[key] = fileData[key]
                }
                fileObj.uploaded = true
                newFiles.push(fileObj)
              }
              data.form.items[i].files = newFiles
            }
          }
          // get files
          for (let i = 0; i < data.form.items.length; i++) {
            for (let j = 0; j < data.form.items[i].files.length; j++) {
              const fileObj = data.form.items[i].files[j]
              if (fileObj.uploaded && (this.checkImageType(fileObj.type) || this.checkVideoType(fileObj.type))) {
                if (checkDefined(fileObj.originalSrc)) {
                  fileObj.src = fileObj.originalSrc
                  delete fileObj.originalSrc
                } else {
                  this.getFileURLRequest(i, j)
                }
              }
            }
          }
          this.updatesAccessToken = data.form.updatesAccessToken
          this.items = data.form.items
          this.multiple = data.form.multiple
          this.files = data.form.files
          this.loading = false
          this.$forceUpdate()
          if (!this.preview) {
            this.createSubscription()
          }
        }).catch(err => {
          console.log(err.message)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    }
  },
  methods: {
    checkImageType(type) {
      return /^image\/.*$/.test(type)
    },
    checkVideoType(type) {
      return /^video\/.*$/.test(type)
    },
    createSubscription() {
      const updateFunction = this.handleUpdates
      this.$apollo.subscribe({
        query: gql`subscription updates {
          formUpdates(id: "${this.formId}", updatesAccessToken: "${this.updatesAccessToken}") {
            id
            name
            items{
              question
              type
              options
              text
              required
              files
              updateAction
              index
              newIndex
            }
            multiple
            files{
              id
              name
              type
              updateAction
              fileIndex
              itemIndex
            }
          }
        }`,
        variables: {}
      }).subscribe({
        next(data) {
          updateFunction(data)
        },
        error(error) {
          const message = `got error: ${error.message}`
          this.$toasted.global.error({
            message
          })
        }
      })
    },
    handleUpdates(data) {
      let foundUpdate = false
      const updateData = data.data.formUpdates
      if (!updateData) return
      if (!this.connectionId && updateData.id) {
        const splitConnectionData = updateData.id.split('connection-')
        if (splitConnectionData.length === 2) {
          this.connectionId = splitConnectionData[1]
        } else {
          return
        }
      }
      if (updateData.name) {
        this.name = updateData.name
        foundUpdate = true
      }
      if (updateData.multiple) {
        this.multiple = updateData.multiple
        foundUpdate = true
      }
      if (checkDefined(updateData.items)) {
        foundUpdate = true
        updateData.items.forEach(item => {
          if (item.updateAction === 'add') {
            const newItem = clone(defaultItem)
            if (newItem.type === itemTypes[0] ||
                newItem.type === itemTypes[1]) {
              newItem.options.push('')
            }
            this.items.push(newItem)
          } else if (item.updateAction === 'set') {
            const index = item.index
            const newItem = clone(defaultItem)
            delete item.updateAction
            delete item.index
            delete item.newIndex
            for (const key in item) {
              if (key === 'files' && (!item.files || item.files.length === 0))
                continue
              newItem[key] = item[key]
            }
            this.items[index] = newItem
          } else if (item.updateAction === 'move') {
            const from = item.index
            const to = item.newIndex
            arrayMove(this.items, from, to)
          } else if (item.updateAction === 'remove') {
            const itemIndex = item.index
            this.items.splice(itemIndex, 1)
          }
        })
      }
      if (checkDefined(updateData.files)) {
        foundUpdate = true
        updateData.files.forEach(file => {
          const itemIndex = file.itemIndex
          if (!this.items[itemIndex])
            return
          const fileIndex = file.fileIndex
          const updateAction = file.updateAction
          delete file.itemIndex
          delete file.fileIndex
          delete file.updateAction
          if (!this.items[itemIndex].files) {
            this.items[itemIndex].files = [clone(defaultFile)]
          } else if (typeof this.items[itemIndex].files[0] === 'number') {
            this.items[itemIndex].files[0] = clone(defaultFile)
          }
          if (updateAction === 'add') {
            while (this.items[itemIndex].files.length < fileIndex) {
              this.items[itemIndex].files.push(clone(defaultFile))
            }
            if (typeof this.items[itemIndex].files[fileIndex] === 'number') {
              this.items[itemIndex].files[fileIndex] = clone(defaultFile)
            }
            for (const key in file) {
              this.items[itemIndex].files[fileIndex][key] = file[key]
            }
          } else if (updateAction === 'set') {
            for (const key in file) {
              this.items[itemIndex].files[fileIndex][key] = file[key]
            }
          } else if (updateAction === 'remove') {
            if (!this.items[itemIndex]) {
              return
            }
            if (this.items[itemIndex].files.length === 1) {
              this.items[itemIndex].files[0] = clone(defaultFile)
            } else {
              this.items[itemIndex].files.splice(fileIndex, 1)
            }
          }
          if (updateAction === 'add' || updateAction === 'set') {
            this.items[itemIndex].files[fileIndex].uploaded = true
            if (this.checkImageType(file.type) || this.checkVideoType(file.type)) {
              if (!this.items[itemIndex].files[fileIndex].src) {
                this.getFileURLRequest(itemIndex, fileIndex)
              }
            }
          }
        })
      }
      if (foundUpdate) {
        this.$forceUpdate()
      }
    },
    getFileURLRequest(itemIndex, fileIndex) {
      // update file src
      this.$axios.get('/getFile', {
        params: {
          posttype: 'form',
          postid: this.formId,
          fileid: this.items[itemIndex].files[fileIndex].id,
          requestType: 'original',
          fileType: this.items[itemIndex].files[fileIndex].type,
          updateToken: this.updatesAccessToken
        }
      }).then(res => {
        if (res.data.url) {
          this.items[itemIndex].files[fileIndex].src = res.data.url
          this.$forceUpdate()
        } else {
          const message = 'cannot find src url in response'
          this.$toasted.global.error({
            message
          })
        }
      }).catch(err => {
        let message = `got error on file url get: ${err}`
        if (err.response && err.response.data) {
          message = err.response.data.message
        }
        this.$toasted.global.error({
          message
        })
      })
    },
    getFileURL(itemIndex, fileIndex) {
      if (this.items[itemIndex].files[fileIndex].src) {
        return this.items[itemIndex].files[fileIndex].src
      } else {
        return ''
      }
    },
    focusItem(evt, itemIndex) {
      this.focusIndex = itemIndex
    },
    submit(evt) {
      evt.preventDefault()
      console.log('submitted!')
    }
  }
})
</script>

<style lang="scss">
.separate-items {
  margin-top: 2rem;
  margin-bottom: 2rem;
}
.submit-button {
  height: 6rem;
  width: 6rem;
  text-align: center;
  line-height: 50%;
}
.card-data {
  @media (min-width: 600px) {
    max-width: 50rem;
  }
}
</style>
