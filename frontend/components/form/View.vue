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
                <b-container v-if="item.type === itemTypes[3]">
                  <p v-html="item.text" />
                </b-container>
                <b-container v-else-if="item.type === itemTypes[7]">
                  <b-img
                    v-if="
                      item.files[0].src &&
                        item.files[0].type &&
                        checkImageType(item.files[0].type)
                    "
                    :src="item.files[0].src"
                    class="mt-2 mb-2 sampleimage"
                  />
                  <video
                    v-else-if="
                      item.files[0].src &&
                        item.files[0].type &&
                        checkVideoType(item.files[0].type)
                    "
                    :ref="`video-source-${index}-${0}`"
                    :type="item.files[0].type"
                    :src="item.files[0].src"
                    controls
                    autoplay
                    class="mb-2 sampleimage"
                    allowfullscreen
                  />
                </b-container>
                <div v-else-if="item.type === itemTypes[6]" class="mt-2 mb-2">
                  <b-container>
                    <b-row>
                      <b-col>
                        <a
                          v-if="items[index].file"
                          :href="getFileURL(index, 0, false)"
                          :download="items[index].files[0].name"
                          class="mt-2 mb-2"
                          >Download</a
                        >
                      </b-col>
                    </b-row>
                  </b-container>
                </div>
                <b-container v-else>
                  <h4>{{ item.question }}</h4>
                  <b-input-group :id="`item-${index}-question`">
                    <b-container>
                      <b-form-radio-group
                        v-if="item.type === itemTypes[0]"
                        :id="`item-${index}-radio`"
                        v-model="item.responseItem.options[0]"
                        :options="item.options"
                        @change="changedResponse(index)"
                        :disabled="!canWrite"
                        name="radios-stacked"
                        stacked
                      />
                      <b-form-checkbox-group
                        v-else-if="item.type === itemTypes[1]"
                        :id="`item-${index}-checkbox`"
                        v-model="item.responseItem.options"
                        :options="item.options"
                        @change="changedResponse(index)"
                        :disabled="!canWrite"
                        name="radios-stacked"
                        stacked
                      />
                      <b-form-textarea
                        v-else-if="item.type === itemTypes[2]"
                        :id="`item-${index}-shortAnswer`"
                        v-model="item.responseItem.text"
                        @change="changedResponse(index)"
                        :disabled="!canWrite"
                        class="mt-2 mb-2"
                        rows="3"
                        max-rows="8"
                        style="max-width:30rem;"
                      />
                      <b-form-checkbox
                        v-else-if="item.type === itemTypes[4]"
                        v-model="item.responseItem.options"
                        :id="`item-${index}-red-green`"
                        @change="changedResponse(index)"
                        :disabled="!canWrite"
                        class="mt-2 mb-2"
                        style="display: inline-block;"
                        name="red-green"
                        switch
                      />
                      <b-form-file
                        v-else-if="item.type === itemTypes[5]"
                        :id="`item-${index}-file-upload`"
                        v-model="item.responseItem.files[0].file"
                        @input="updateFileSrc(index, 0)"
                        @change="changedResponse(index)"
                        :disabled="!canWrite"
                        class="mt-2 mb-2"
                        placeholder="Choose a file or drop it here..."
                        drop-placeholder="Drop file here..."
                        style="max-width:30rem;"
                      />
                    </b-container>
                  </b-input-group>
                </b-container>
              </span>
            </div>
          </b-form>
        </b-card-body>
      </b-card>
      <b-container
        v-if="!preview && canWrite"
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
import { noneAccessType } from '~/assets/config'
import { clone, arrayMove, checkDefined } from '~/assets/utils'
const itemTypes = ['radio', 'checkbox', 'short', 'text',
  'redgreen', 'fileupload', 'fileattachment', 'media']
const responseItemTypes = [itemTypes[0], itemTypes[1], itemTypes[2],
  itemTypes[4], itemTypes[5]]
const objectType = 'response'
const defaultFile = {
  id: '',
  name: '',
  type: '',
  src: null,
  uploaded: false,
  uploadProgress: 0,
  updateAction: 'add'
}
const defaultResponseItem = {
  text: '',
  options: [],
  files: [clone(defaultFile)],
  uploaded: false,
  updateAction: 'add'
}
const defaultItem = {
  question: '',
  type: itemTypes[0],
  options: [],
  text: '',
  required: false,
  files: [clone(defaultFile)],
  responseItem: null
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
    responseId: {
      type: String,
      default: null
    },
    accessToken: {
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
      numResponses: 0,
      items: [],
      multiple: false,
      focusIndex: 0,
      itemTypes,
      loading: true,
      isPublic: false,
      updatesAccessToken: null,
      connectionId: null,
      userIdResponse: null,
      editResponseAccessToken: null
    }
  },
  computed: {
    canWrite() {
      return (this.multiple || this.numResponses === 0) && (!this.userIdResponse || this.userIdResponse === this.$store.state.auth.user.id)
    }
  },
  mounted() {
    this.currentResponseId = this.responseId
    if (this.formId) {
      this.$apollo.query({
        query: gql`
          query form($id: String!, $accessToken: String){
            form(id: $id, editAccessToken: false, accessToken: $accessToken) {
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
              ${!this.currentResponseId ? 'updatesAccessToken' : ''}
            }
          }`,
          variables: {id: this.formId, accessToken: this.accessToken},
          fetchPolicy: 'network-only'
        })
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
            const currentItem = data.form.items[i]
            for (let j = 0; j < currentItem.files.length; j++) {
              const fileObj = currentItem.files[j]
              if (fileObj.uploaded && (this.checkImageType(fileObj.type) || this.checkVideoType(fileObj.type))) {
                if (checkDefined(fileObj.originalSrc)) {
                  fileObj.src = fileObj.originalSrc
                  delete fileObj.originalSrc
                } else {
                  this.getFileURLRequest(i, j, false)
                }
              }
            }
            if (responseItemTypes.includes(currentItem.type)) {
              currentItem.responseItem = clone(defaultResponseItem)
            }
          }
          this.updatesAccessToken = data.form.updatesAccessToken
          this.items = data.form.items
          this.multiple = data.form.multiple
          if (!this.preview && !this.currentResponseId) {
            this.createSubscription()
          }
          if (!this.preview && this.currentResponseId) {
            this.getResponseData()
          } else {
            this.loading = false
            this.$forceUpdate()
          }
        }).catch(err => {
          console.error(err.message)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    }
  },
  methods: {
    getResponseData() {
      if (this.currentResponseId) {
          this.$apollo.query({
            query: gql`
              query response($id: String!) {
                response(id: $id) {
                  user
                  items {
                    formIndex
                    text
                    options
                    files
                  }
                  files {
                    id
                    name
                    type
                    originalSrc
                  }
                }
              }`,
              variables: {id: this.currentResponseId},
              fetchPolicy: 'network-only'
            }).then(({ data }) => {
              console.log(data.response)
              this.userIdResponse = data.response.user
              for (let i = 0; i < data.response.items.length; i++) {
                if (data.response.items[i].files.length === 0) {
                  data.response.items[i].files = [clone(defaultFile)]
                } else {
                  const newFiles = []
                  for (let j = 0; j < data.response.items[i].files.length; j++) {
                    const fileData = data.response.files[data.response.items[i].files[j]]
                    const fileObj = clone(defaultFile)
                    for (const key in fileData) {
                      fileObj[key] = fileData[key]
                    }
                    fileObj.uploaded = true
                    newFiles.push(fileObj)
                  }
                  data.response.items[i].files = newFiles
                }
              }
              // get files
              for (let i = 0; i < data.response.items.length; i++) {
                for (let j = 0; j < data.response.items[i].files.length; j++) {
                  const fileObj = data.response.items[i].files[j]
                  if (fileObj.uploaded) {
                    if (checkDefined(fileObj.originalSrc)) {
                      fileObj.src = fileObj.originalSrc
                      delete fileObj.originalSrc
                    } else {
                      this.getFileURLRequest(i, j, true)
                    }
                    fileObj.updateAction = 'set'
                  }
                }
              }
              for (let i = 0; i < data.response.items.length; i++) {
                const currentObj = data.response.items[i]
                const itemIndex = currentObj.formIndex
                delete currentObj.formIndex
                currentObj.uploaded = true
                currentObj.updateAction = 'set'
                this.items[itemIndex].responseItem = currentObj
              }
              console.log('done getting items')
              console.log(this.items[0])
              this.loading = false
              this.$forceUpdate()
            }).catch(err => {
              console.log(err.message)
              this.$toasted.global.error({
                message: `found error: ${err.message}`
              })
            })
      }
    },
    checkImageType(type) {
      return /^image\/.*$/.test(type)
    },
    checkVideoType(type) {
      return /^video\/.*$/.test(type)
    },
    changedResponse(itemIndex) {
      if (this.items[itemIndex].responseItem.uploaded) {
        this.items[itemIndex].responseItem.uploaded = false
      }
    },
    updateFileSrc(itemIndex, fileIndex, justUploaded) {
      const fileObj = this.items[itemIndex].responseItem.files[fileIndex]
      if (fileObj.file && !fileObj.src) {
        if (!fileObj.type) {
          fileObj.type = fileObj.file.type
        }
        if (!fileObj.name) {
          fileObj.name = fileObj.file.name
        }
      }
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
                this.getFileURLRequest(itemIndex, fileIndex, false)
              }
            }
          }
        })
      }
      if (foundUpdate) {
        this.$forceUpdate()
      }
    },
    getFileURLRequest(itemIndex, fileIndex, isResponse) {
      const updateObj = isResponse ? this.items[itemIndex] : this.items[itemIndex].responseItem
      // update file src
      this.$axios.get('/getFile', {
        params: {
          posttype: 'form',
          postid: this.formId,
          fileid: updateObj.files[fileIndex].id,
          requestType: 'original',
          fileType: updateObj.files[fileIndex].type,
          updateToken: isResponse ? this.editResponseAccessToken : this.updatesAccessToken
        }
      }).then(res => {
        if (res.data.url) {
          updateObj.files[fileIndex].src = res.data.url
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
    getFileURL(itemIndex, fileIndex, isResponse) {
      const dataObj = isResponse ? this.items[itemIndex] : this.items[itemIndex].responseItem
      if (dataObj.files[fileIndex].src) {
        return dataObj.files[fileIndex].src
      } else {
        return ''
      }
    },
    focusItem(evt, itemIndex) {
      this.focusIndex = itemIndex
    },
    getTotalResponseFileIndex(itemIndex, fileIndex) {
      let totalFileIndex = 0
      for (let i = 0; i < itemIndex; i++) {
        this.items[i].responseItem.files.forEach(file => {
          if (file.uploaded) {
            totalFileIndex++
          }
        })
      }
      return totalFileIndex + fileIndex
    },
    getResponseItemData(itemIndex) {
      const item = {
        text: this.items[itemIndex].responseItem.text,
        options: this.items[itemIndex].responseItem.options,
        files: [],
        formIndex: itemIndex,
        updateAction: this.items[itemIndex].responseItem.updateAction
      }
      let currentFileIndex = this.getTotalResponseFileIndex(itemIndex, 0)
      for (let i = 0; i < this.items[itemIndex].files.length; i++) {
        if (this.items[itemIndex].responseItem.files[i].uploaded) {
          item.files.push(currentFileIndex)
          currentFileIndex++
        }
      }
      return item
    },
    submit(evt) {
      evt.preventDefault()
      let remainingFileOperations = 0
      for (let i = 0; i < this.items.length; i++) {
        if (this.items[i].type !== itemTypes[5]) continue
        for (let j = 0; j < this.items[i].responseItem.files.length; j++) {
          if (this.items[i].responseItem.files[j].file) {
            remainingFileOperations++
          }
        }
      }
      const onFileUploadComplete = () => {
        const uploadItems = []
        const uploadFiles = []
        let responseCount = 0
        for (let i = 0; i < this.items.length; i++) {
          if (!responseItemTypes.includes(this.items[i].type)) continue
          if (!this.items[i].responseItem.uploaded) {
            const currentResponse = this.getResponseItemData(i)
            if (this.currentResponseId) {
              currentResponse.index = responseCount
            }
            uploadItems.push(currentResponse)
            for (let j = 0; j < this.items[i].responseItem.files.length; j++) {
              if (!this.items[i].responseItem.files[j].file) continue
              const fileObj = this.items[i].responseItem.files[j]
              uploadFiles.push({
                updateAction: fileObj.updateAction,
                index: this.getTotalResponseFileIndex(i, j),
                id: fileObj.id,
                itemIndex: i,
                fileIndex: j
              })
            }
            this.items[i].responseItem.uploaded = true
          }
          responseCount++
        }
        console.log(uploadItems)
        this.$apollo.mutate({mutation: gql`
          mutation updateResponse($accessToken: String, $id: String!, $items: [UpdateResponseItemInput!]!, $files: [UpdateFileInput!]!)
          {updateResponse(accessToken: $accessToken, id: $id, items: $items, files: $files){id} }
          `, variables: {id: this.currentResponseId, items: uploadItems, files: uploadFiles, accessToken: this.editResponseAccessToken}})
          .then(({ data }) => {
            console.log('submitted!')
          }).catch(err => {
            console.error(err)
            this.$toasted.global.error({
              message: `found error: ${err.message}`
            })
          })
      }
      const uploadFiles = () => {
        let foundFile = false
        for (let i = 0; i < this.items.length; i++) {
          if (!this.items[i].responseItem) continue
          for (let j = 0; j < this.items[i].responseItem.files.length; j++) {
            const fileObj = this.items[i].files[j]
            if (!fileObj.file) continue
            foundFile = true
            fileObj.uploadProgress = 0
            if (fileObj.uploaded) {
              remainingFileOperations++
              this.$axios
                .delete('/deleteFiles', {
                  data: {
                    fileids: [
                      fileObj.id
                    ],
                    postid: this.currentResponseId,
                    posttype: objectType,
                    updateToken: this.editResponseAccessToken
                  }
                })
                .then(res => {
                  if (res.status === 200) {
                    console.log('deleted file')
                    remainingFileOperations--
                    if (remainingFileOperations === 0) {
                      onFileUploadComplete()
                    }
                  } else {
                    this.$toasted.global.error({
                      message: `got status code of ${res.status} on file delete`
                    })
                  }
                })
                .catch(err => {
                  let message = `got error on file delete: ${err}`
                  if (err.response && err.response.data) {
                    message = err.response.data.message
                  }
                  this.$toasted.global.error({
                    message
                  })
                })
            }
            const formData = new FormData()
            formData.append('file', fileObj.file)
            this.$axios
              .put('/writeFile', formData, {
                params: {
                  posttype: objectType,
                  filetype: fileObj.file.type,
                  postid: this.currentResponseId,
                  updateToken: this.editResponseAccessToken
                },
                headers: {
                  'Content-Type': 'multipart/form-data'
                }
              })
              .then(res => {
                if (res.status === 200) {
                  fileObj.uploadProgress = 100
                  fileObj.uploaded = true
                  fileObj.id = res.data.id
                  this.updateFileSrc(i, j, true)
                  remainingFileOperations--
                  if (remainingFileOperations === 0) {
                    onFileUploadComplete()
                  }
                } else {
                  this.$toasted.global.error({
                    message: `got status code of ${res.status} on file upload`
                  })
                }
              })
              .catch(err => {
                let message = `got error: ${err}`
                if (err.response && err.response.data) {
                  message = err.response.data.message
                }
                this.$toasted.global.error({
                  message
                })
              })
          }
        }
        if (!foundFile) {
          onFileUploadComplete()
        }
      }
      if (!this.currentResponseId) {
        console.log(`form id: ${this.formId}`)
        this.$apollo.mutate({mutation: gql`
          mutation addResponse($accessToken: String, $id: String!, $items: [ResponseItemInput!]!, $files: [FileInput!]!)
          {addResponse(accessToken: $accessToken, id: $id, items: $items, files: $files){id editAccessToken} }
          `, variables: {accessToken: this.accessToken, id: this.formId, items: [], files: []}})
          .then(({ data }) => {
            this.currentResponseId = data.addResponse.id
            this.userIdResponse = this.$store.state.auth.user.id
            this.editResponseAccessToken = data.addResponse.editAccessToken
            history.replaceState({}, null, `/project/${this.projectId}/form/${this.formId}/response/${this.currentResponseId}`)
            uploadFiles()
          }).catch(err => {
            console.error(err)
            this.$toasted.global.error({
              message: `found error: ${err.message}`
            })
          })
      } else {
        uploadFiles()
      }
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
