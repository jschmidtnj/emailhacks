<template>
  <div id="create">
    <div v-if="!loading">
      <b-card no-body class="card-data shadow-lg">
        <b-card-body>
          <b-form @submit.prevent>
            <b-input-group>
              <b-container>
                <b-row class="mb-2">
                  <b-col sm>
                    <b-form-input
                      id="name"
                      v-model="name"
                      @change="setUpdates({ main: ['name'] })"
                      size="lg"
                      type="text"
                      placeholder="Title"
                    />
                  </b-col>
                </b-row>
                <b-row class="mt-2 mb-2">
                  <b-col sm>
                    <b-form-checkbox
                      v-model="multiple"
                      @change="setUpdates({ main: ['multiple'] })"
                      class="pull-right"
                      name="allow-multiple"
                      switch
                    >
                      Allow Multiple Submissions
                    </b-form-checkbox>
                  </b-col>
                </b-row>
              </b-container>
            </b-input-group>
            <hr />
            <draggable
              v-model="items"
              @end="(evt) => finishedDragging(evt, true)"
              group="items"
              ghost-class="ghost"
            >
              <div
                v-for="(item, index) in items"
                :key="`item-${index}`"
                :class="{ 'item-focus': focusIndex === index }"
              >
                <span
                  :id="`item-${index}-select-area`"
                  v-touch:start="(evt) => focusItem(evt, index)"
                >
                  <div class="drag-area">
                    <client-only>
                      <font-awesome-icon
                        class="icon-grip"
                        icon="grip-horizontal"
                      />
                    </client-only>
                  </div>
                  <b-input-group :id="`item-${index}-name-type-input`">
                    <b-container>
                      <b-row>
                        <b-col sm class="my-auto">
                          <b-form-input
                            v-if="item.type !== itemTypes[3].id"
                            :id="`item-${index}-name`"
                            v-model="item.question"
                            @change="
                              setUpdates({
                                items: [
                                  {
                                    updateAction: 'set',
                                    index,
                                    ...getItemData(index)
                                  }
                                ]
                              })
                            "
                            size="md"
                            type="text"
                            placeholder="Question"
                          />
                          <div v-else />
                        </b-col>
                        <b-col sm>
                          <b-dropdown
                            v-if="focusIndex === index"
                            :id="`item-type-${index}`"
                            :text="getItemTypeLabel(index)"
                            variant="outline-primary"
                            class="mt-2 mb-2"
                          >
                            <b-dropdown-item-button
                              v-for="(type, indexType) in itemTypes"
                              :key="`item-${index}-select-${indexType}`"
                              @click="(evt) => selectItemType(evt, index, type)"
                              >{{ type.label }}</b-dropdown-item-button
                            >
                          </b-dropdown>
                        </b-col>
                      </b-row>
                    </b-container>
                  </b-input-group>
                  <b-input-group :id="`item-${index}-content`">
                    <b-container>
                      <div
                        v-if="
                          item.type === itemTypes[0].id ||
                            item.type === itemTypes[1].id
                        "
                      >
                        <b-row
                          v-for="(option, optionIndex) in item.options"
                          :key="`item-${index}-option-${optionIndex}`"
                          class="mt-2 mb-2"
                          style="max-width:30rem;"
                        >
                          <b-col style="max-width:30px;">
                            <b-form-radio
                              v-if="item.type === itemTypes[0].id"
                              disabled
                            />
                            <b-form-checkbox
                              v-else-if="item.type === itemTypes[1].id"
                              disabled
                            />
                          </b-col>
                          <b-col
                            v-if="
                              item.type === itemTypes[0].id ||
                                item.type === itemTypes[1].id
                            "
                          >
                            <b-form-input
                              v-model="item.options[optionIndex]"
                              :placeholder="`option ${optionIndex + 1}`"
                              @change="
                                setUpdates({
                                  items: [
                                    {
                                      updateAction: 'set',
                                      index,
                                      ...getItemData(index)
                                    }
                                  ]
                                })
                              "
                              size="sm"
                              type="text"
                              style="width:100%;"
                            />
                          </b-col>
                          <b-col style="padding-left:0;max-width:30px;">
                            <button
                              :disabled="item.options.length <= 1"
                              :class="{
                                'disable-button': item.options.length <= 1
                              }"
                              @click="
                                (evt) => removeOption(evt, index, optionIndex)
                              "
                              class="button-link"
                            >
                              <client-only>
                                <font-awesome-icon class="mr-2" icon="times" />
                              </client-only>
                            </button>
                          </b-col>
                        </b-row>
                        <b-row
                          v-if="
                            item.type === itemTypes[0].id ||
                              item.type === itemTypes[1].id
                          "
                        >
                          <b-col style="max-width:30px;">
                            <b-form-radio
                              v-if="item.type === itemTypes[0].id"
                              disabled
                            />
                            <b-form-checkbox v-else disabled />
                          </b-col>
                          <b-col>
                            <button
                              :disabled="
                                item.options.length !== 0 &&
                                  item.options[item.options.length - 1]
                                    .length === 0
                              "
                              :class="{
                                'disable-button':
                                  item.options.length !== 0 &&
                                  item.options[item.options.length - 1]
                                    .length === 0
                              }"
                              @click="(evt) => addOption(evt, index)"
                              class="button-link"
                            >
                              Add
                              {{
                                item.type === itemTypes[0].id
                                  ? 'Multiple Choice'
                                  : 'Checkbox'
                              }}
                              Option
                            </button>
                          </b-col>
                        </b-row>
                      </div>
                      <b-form-input
                        v-else-if="item.type === itemTypes[2].id"
                        :id="`item-${index}-shortAnswer`"
                        size="sm"
                        type="text"
                        disabled
                        placeholder="short answer"
                        class="mt-2 mb-2"
                        style="max-width:30rem;"
                      />
                      <div
                        v-else-if="item.type === itemTypes[3].id"
                        class="editor mt-2 mb-2"
                      >
                        <text-editor
                          :ref="`editor-${index}`"
                          :show-menu="index === focusIndex"
                          @updated-text="(text) => updatedText(text, index)"
                        />
                      </div>
                      <div
                        v-else-if="item.type === itemTypes[4].id"
                        class="mt-2 mb-2"
                      >
                        <b-form-checkbox
                          :id="`item-${index}-red-green`"
                          style="display: inline-block;"
                          name="red-green"
                          switch
                          disabled
                        />
                      </div>
                      <div
                        v-else-if="item.type === itemTypes[5].id"
                        class="mt-2 mb-2"
                      >
                        <b-form-file
                          placeholder="Choose a file or drop it here..."
                          drop-placeholder="Drop file here..."
                          style="max-width:30rem;"
                          disabled
                        />
                      </div>
                      <div
                        v-else-if="item.type === itemTypes[6].id"
                        class="mt-2 mb-2"
                      >
                        <b-container>
                          <b-row>
                            <b-col style="padding-left:0;">
                              <b-form-file
                                :id="`item-${index}-file-attachment`"
                                v-model="item.files[0].file"
                                :accept="validfiles.join(', ')"
                                @input="uploadFile(index, 0)"
                                placeholder="Choose a file or drop it here..."
                                drop-placeholder="Drop file here..."
                                style="max-width:30rem;"
                                class="mb-2"
                              />
                            </b-col>
                          </b-row>
                          <b-row>
                            <b-col>
                              <b-progress
                                v-if="
                                  item.files[0].file && !item.files[0].uploaded
                                "
                                :value="item.files[0].uploadProgress"
                                :max="100"
                                show-progress
                                animated
                              ></b-progress>
                              <a
                                v-else-if="item.files[0].uploaded"
                                :href="getFileURL(index, 0)"
                                :download="items[index].files[0].name"
                                target="_blank"
                                class="mt-2 mb-2"
                                >Download</a
                              >
                            </b-col>
                          </b-row>
                        </b-container>
                      </div>
                      <div
                        v-else-if="item.type === itemTypes[7].id"
                        class="mt-2 mb-2"
                      >
                        <b-container>
                          <b-row>
                            <b-col style="padding-left:0;">
                              <b-form-file
                                :id="`item-${index}-image`"
                                v-model="item.files[0].file"
                                :accept="validDisplayFiles.join(', ')"
                                @input="uploadFile(index, 0)"
                                placeholder="Choose an image or drop it here..."
                                drop-placeholder="Drop image here..."
                                style="max-width:30rem;"
                                class="mb-2"
                              />
                            </b-col>
                          </b-row>
                          <b-row>
                            <b-col style="padding-left:0;">
                              <b-progress
                                v-if="
                                  item.files[0].file && !item.files[0].uploaded
                                "
                                :value="item.files[0].uploadProgress"
                                :max="100"
                                show-progress
                                animated
                              ></b-progress>
                              <b-img
                                v-else-if="
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
                            </b-col>
                          </b-row>
                        </b-container>
                      </div>
                    </b-container>
                  </b-input-group>
                </span>
                <div v-if="focusIndex === index">
                  <hr />
                  <b-input-group>
                    <b-container>
                      <b-row>
                        <b-col class="text-right">
                          <button
                            :disabled="items.length <= 1"
                            @click="(evt) => removeItem(evt, index, true)"
                            class="button-link"
                            style="display: inline-block;"
                          >
                            <client-only>
                              <font-awesome-icon class="mr-2" icon="trash" />
                            </client-only>
                          </button>
                          <b-form-checkbox
                            v-model="item.required"
                            style="display: inline-block;"
                            name="required"
                            switch
                          >
                            Required
                          </b-form-checkbox>
                        </b-col>
                      </b-row>
                    </b-container>
                  </b-input-group>
                </div>
                <div v-else class="separate-items">
                  <hr />
                </div>
              </div>
            </draggable>
            <b-container>
              <b-row>
                <b-col class="text-center">
                  <b-button
                    @click="(evt) => addItem(evt, true)"
                    pill
                    variant="primary"
                    class="add-button"
                  >
                    <client-only>
                      <font-awesome-icon size="lg" icon="plus" />
                    </client-only>
                  </b-button>
                </b-col>
              </b-row>
            </b-container>
          </b-form>
        </b-card-body>
      </b-card>
      <b-container style="margin-top: 3rem; margin-bottom: 2rem;">
        <b-row>
          <b-col class="text-right">
            <b-button
              @click="send"
              pill
              variant="primary"
              class="send-button shadow-lg"
            >
              <client-only>
                <font-awesome-icon size="3x" icon="paper-plane" />
              </client-only>
            </b-button>
          </b-col>
        </b-row>
      </b-container>
      <send-modal ref="send-content-modal" :form-id="formId" />
    </div>
    <page-loading v-else :loading="true" />
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import Draggable from 'vuedraggable'
import TextEditor from '~/components/form/TextEditor.vue'
import SendModal from '~/components/form/SendModal.vue'
import PageLoading from '~/components/PageLoading.vue'
import { validfiles, validDisplayFiles, autosaveInterval } from '~/assets/config'
import { clone, arrayMove, checkDefined } from '~/assets/utils'
const objectType = 'form'
// still need image picker and image viewer component
const itemTypes = [
  {
    id: 'radio',
    label: 'Multiple Choice'
  },
  {
    id: 'checkbox',
    label: 'Checkbox'
  },
  {
    id: 'short',
    label: 'Short Answer'
  },
  {
    id: 'text',
    label: 'Text'
  },
  {
    id: 'redgreen',
    label: 'Red / Green Light'
  },
  {
    id: 'fileupload',
    label: 'File Upload'
  },
  {
    id: 'fileattachment',
    label: 'File Attachment'
  },
  {
    id: 'media',
    label: 'Media'
  },
]
const defaultFile = {
  id: '',
  name: '',
  width: null,
  height: null,
  type: '',
  file: null,
  src: null,
  uploadProgress: 0,
  uploaded: false
}
const defaultItem = {
  question: '',
  type: itemTypes[0].id,
  options: [],
  text: '',
  required: false,
  files: [clone(defaultFile)]
}
const defaultPendingUpdates = {
  main: new Set(),
  items: [],
  files: []
}
export default Vue.extend({
  name: 'Create',
  components: {
    TextEditor,
    PageLoading,
    SendModal,
    Draggable
  },
  props: {
    formId: {
      type: String,
      default: null
    }
  },
  data() {
    return {
      validDisplayFiles,
      validfiles,
      loading: true,
      itemTypes,
      updateTimer: null,
      pendingUpdates: clone(defaultPendingUpdates),
      focusIndex: 0,
      editorContent: {},
      name: '',
      items: [],
      multiple: false,
      updatesAccessToken: null,
      connectionId: null
    }
  },
  mounted() {
    this.$apollo.query({
      query: gql`
        query form($id: String!) {
          form(id: $id, editAccessToken: true) {
            name
            items{
              question
              type
              options
              text
              required
              files
            }
            multiple
            files{
              id
              name
              width
              height
              type
              originalSrc
            }
            updatesAccessToken
          }
        }`,
        variables: {id: this.formId},
        fetchPolicy: 'network-only'
      }).then(({ data }) => {
        this.name = data.form.name
        const newEditorContent = {}
        for (let i = 0; i < data.form.items.length; i++) {
          if (data.form.items[i].type === itemTypes[3].id) {
            newEditorContent[i] = data.form.items[i].text
          }
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
        this.updatesAccessToken = data.form.updatesAccessToken
        this.editorContent = newEditorContent
        // get files
        for (let i = 0; i < data.form.items.length; i++) {
          for (let j = 0; j < data.form.items[i].files.length; j++) {
            const fileObj = data.form.items[i].files[j]
            if (fileObj.uploaded && (this.checkImageType(fileObj.type) || this.checkVideoType(fileObj.type))) {
              if (fileObj.originalSrc) {
                fileObj.src = fileObj.originalSrc
                delete fileObj.originalSrc
              } else {
                this.getFileURLRequest(i, j)
              }
            }
          }
        }
        this.items = data.form.items
        this.multiple = data.form.multiple
        this.loading = false
        this.$forceUpdate()
        this.createSubscription()
        this.$nextTick(() => {
          const newTextLocations = Object.keys(this.editorContent)
          for (let i = 0; i < newTextLocations.length; i++) {
            if (this.$refs[`editor-${newTextLocations[i]}`]) {
              this.$refs[`editor-${newTextLocations[i]}`][0]._data.editor.setContent(this.editorContent[newTextLocations[i]])
            }
          }
        })
        // save before leave window
      }).catch(err => {
        this.$bvToast.toast(`found error: ${err.message}`, {
          variant: 'danger',
          title: 'Error'
        })
      })
  },
  methods: {
    send(evt) {
      evt.preventDefault()
      console.log('send!')
      if (this.$refs['send-content-modal']) {
        this.$refs['send-content-modal'].show()
      } else {
        this.$bvToast.toast('cannot find send modal', {
          variant: 'danger',
          title: 'Error'
        })
      }
    },
    getFileData(itemIndex, fileIndex) {
      return {
        fileIndex,
        itemIndex,
        id: this.items[itemIndex].files[fileIndex].id,
        name: this.items[itemIndex].files[fileIndex].name,
        width: this.items[itemIndex].files[fileIndex].width,
        height: this.items[itemIndex].files[fileIndex].height,
        type: this.items[itemIndex].files[fileIndex].type
      }
    },
    getTotalFileIndex(itemIndex, fileIndex) {
      let totalFileIndex = 0
      for (let i = 0; i < itemIndex; i++) {
        this.items[i].files.forEach(file => {
          if (file.uploaded) {
            totalFileIndex++
          }
        })
      }
      return totalFileIndex + fileIndex
    },
    getItemData(itemIndex) {
      const item = {
        question: this.items[itemIndex].question,
        type: this.items[itemIndex].type,
        options: this.items[itemIndex].options,
        text: '',
        required: this.items[itemIndex].required,
        files: []
      }
      if (item.type === itemTypes[3].id) {
        this.items[itemIndex].text = this.editorContent[itemIndex]
        item.text = this.items[itemIndex].text
      }
      let currentFileIndex = this.getTotalFileIndex(itemIndex, 0)
      for (let i = 0; i < this.items[itemIndex].files.length; i++) {
        if (this.items[itemIndex].files[i].uploaded) {
          item.files.push(currentFileIndex)
          currentFileIndex++
        }
      }
      return item
    },
    setUpdates({ main, items, files }) {
      if (main) {
        main.forEach(update => {
          this.pendingUpdates.main.add(update)
        })
      }
      if (items) {
        items.forEach(item => {
          if (item.updateAction === 'set') {
            const itemIndex = this.pendingUpdates.items.findIndex(currentItem => {
              if (currentItem.updateAction !== 'set') return false
              return item.index === currentItem.index
            })
            if (itemIndex < 0) {
              this.pendingUpdates.items.push(item)
            } else {
              this.pendingUpdates.items[itemIndex] = item
            }
          } else {
            this.pendingUpdates.items.push(item)
          }
        })
      }
      if (files) {
        files.forEach(file => {
          if (file.updateAction === 'set') {
            const fileIndex = this.pendingUpdates.files.findIndex(currentFile => {
              if (currentFile.updateAction !== 'set') return false
              return file.id === currentFile.id
            })
            if (fileIndex < 0) {
              this.pendingUpdates.files.push(file)
            } else {
              this.pendingUpdates.files[fileIndex] = file
            }
          } else {
            this.pendingUpdates.files.push(file)
          }
        })
      }
      if (!this.updateTimer) {
        this.updateTimer = setTimeout(this.update, autosaveInterval)
      }
    },
    update() {
      this.updateTimer = null
      this.$apollo.mutate({mutation: gql`
        mutation updateFormPart($id: String!, $updatesAccessToken: String!, $name: String, $items: [UpdateFormItemInput!], $multiple: Boolean, $files: [UpdateFileInput!]) {
          updateFormPart(id: $id, updatesAccessToken: $updatesAccessToken, name: $name, items: $items, multiple: $multiple, files: $files) {
            id
          }
        }
        `, variables: {
          id: this.formId,
          updatesAccessToken: this.updatesAccessToken,
          name: this.pendingUpdates.main.has('name') ? this.name : null,
          multiple: this.pendingUpdates.main.has('multiple') ? this.multiple : null,
          items: this.pendingUpdates.items.length > 0 ? this.pendingUpdates.items : null,
          files: this.pendingUpdates.files.length > 0 ? this.pendingUpdates.files : null
        }})
        .then(({ data }) => {
          this.pendingUpdates = clone(defaultPendingUpdates)
        }).catch(err => {
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
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
      if (updateData.id === this.connectionId) return
      if (checkDefined(updateData.name)) {
        this.name = updateData.name
        foundUpdate = true
      }
      if (checkDefined(updateData.multiple)) {
        this.multiple = updateData.multiple
        foundUpdate = true
      }
      if (checkDefined(updateData.items)) {
        foundUpdate = true
        updateData.items.forEach(item => {
          if (item.updateAction === 'add') {
            this.addItem(null, false)
          } else if (item.updateAction === 'set') {
            const index = item.index
            const newItem = clone(defaultItem)
            delete item.updateAction
            delete item.index
            delete item.newIndex
            let foundFiles = false
            for (const key in item) {
              if (key === 'files' && this.items[index].files && this.items[index].files.length > 0) {
                foundFiles = true
              } else {
                newItem[key] = item[key]
              }
            }
            if (foundFiles) {
              newItem.files = this.items[index].files
            }
            this.items[index] = newItem
            if (this.items[index].type === itemTypes[3].id) {
              this.editorContent[index] = this.items[index].text
              this.$nextTick(() => {
                this.updateEditorContent(index)
              })
            }
          } else if (item.updateAction === 'move') {
            const from = item.index
            const to = item.newIndex
            arrayMove(this.items, from, to)
            this.$nextTick(() => {
              this.finishedDragging({
                oldIndex: from,
                newIndex: to
              }, false)
            })
          } else if (item.updateAction === 'remove') {
            const index = item.index
            this.removeItem(null, index, false)
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
              this.getFileURLRequest(itemIndex, fileIndex)
            }
          }
        })
      }
      if (foundUpdate) {
        this.$forceUpdate()
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
              width
              height
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
          this.$bvToast.toast(message, {
            variant: 'danger',
            title: 'Error'
          })
        }
      })
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
          this.$nextTick(() => {
            this.$forceUpdate()
          })
        } else {
          const message = 'cannot find src url in response'
          this.$bvToast.toast(message, {
            variant: 'danger',
            title: 'Error'
          })
        }
      }).catch(err => {
        let message = `got error on file url get: ${err}`
        if (err.response && err.response.data) {
          message = err.response.data.message
        }
        this.$bvToast.toast(message, {
          variant: 'danger',
          title: 'Error'
        })
      })
    },
    getFileURL(itemIndex, fileIndex) {
      if (this.items[itemIndex].files[fileIndex].uploaded) {
        if (this.items[itemIndex].files[fileIndex].src) {
          return this.items[itemIndex].files[fileIndex].src
        } else {
          return ''
        }
      }
      return URL.createObjectURL(this.items[itemIndex].files[fileIndex].file)
    },
    checkImageType(type) {
      return /^image\/.*$/.test(type)
    },
    checkVideoType(type) {
      return /^video\/.*$/.test(type)
    },
    deleteFile(itemIndex, fileIndex) {
      const fileObj = this.items[itemIndex].files[fileIndex]
      fileObj.src = null
      fileObj.uploaded = false
      this.setUpdates({
        items: [{
          updateAction: 'set',
          index: itemIndex,
          ...this.getItemData(itemIndex)
        }],
        files: [{
          updateAction: 'remove',
          index: this.getTotalFileIndex(itemIndex, fileIndex),
          id: fileObj.id,
          itemIndex,
          fileIndex
        }]
      })
      this.$axios
        .delete('/deleteFiles', {
          data: {
            fileids: [
              fileObj.id
            ],
            postid: this.formId,
            posttype: objectType
          }
        })
        .then(res => {
          if (res.status === 200) {
            console.log('deleted file')
          } else {
            this.$bvToast.toast(`got status code of ${res.status} on file delete`, {
              variant: 'danger',
              title: 'Error'
            })
          }
        })
        .catch(err => {
          let message = `got error on file delete: ${err}`
          if (err.response && err.response.data) {
            message = err.response.data.message
          }
          this.$bvToast.toast(message, {
            variant: 'danger',
            title: 'Error'
          })
        })
    },
    uploadFile(itemIndex, fileIndex) {
      const fileObj = this.items[itemIndex].files[fileIndex]
      if (!fileObj.file) return
      fileObj.uploadProgress = 0
      if (fileObj.uploaded) {
        this.deleteFile(itemIndex, fileIndex)
      }
      const formData = new FormData()
      formData.append('file', fileObj.file)
      this.$axios
        .put('/writeFile', formData, {
          params: {
            posttype: objectType,
            filetype: fileObj.file.type,
            postid: this.formId
          },
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        })
        .then(res => {
          if (res.status === 200) {
            fileObj.name = ''
            fileObj.width = null
            fileObj.height = null
            fileObj.type = ''
            fileObj.uploadProgress = 100
            fileObj.uploaded = true
            fileObj.id = res.data.id
            fileObj.src = null
            this.updateFileSrc(itemIndex, fileIndex, true)
          } else {
            this.$bvToast.toast(`got status code of ${res.status} on file upload`, {
              variant: 'danger',
              title: 'Error'
            })
          }
        })
        .catch(err => {
          let message = `got error: ${err}`
          if (err.response && err.response.data) {
            message = err.response.data.message
          }
          this.$bvToast.toast(message, {
            variant: 'danger',
            title: 'Error'
          })
          fileObj.file = null
        })
    },
    updateFileSrc(itemIndex, fileIndex, justUploaded) {
      const fileObj = this.items[itemIndex].files[fileIndex]
      if (fileObj.file && !fileObj.src) {
        if (!fileObj.type) {
          fileObj.type = fileObj.file.type
        }
        if (!fileObj.name) {
          fileObj.name = fileObj.file.name
        }
        if (this.checkVideoType(fileObj.type)) {
          this.updateVideoSrc(itemIndex, fileIndex, justUploaded)
        } else if (this.checkImageType(fileObj.type)) {
          this.updateImageSrc(itemIndex, fileIndex, justUploaded)
        } else {
          this.$forceUpdate()
          this.setUpdates({
            items: [{
              updateAction: 'set',
              index: itemIndex,
              ...this.getItemData(itemIndex)
            }],
            files: [{
              updateAction: justUploaded ? 'add' : 'set',
              index: this.getTotalFileIndex(itemIndex, fileIndex),
              ...this.getFileData(itemIndex, fileIndex)
            }]
          })
        }
      }
    },
    updateImageSrc(itemIndex, fileIndex, justUploaded) {
      const fileObj = this.items[itemIndex].files[fileIndex]
      if (!fileObj.file) return
      const img = new Image()
      img.onload = () => {
        this.$forceUpdate()
        fileObj.width = img.width
        fileObj.height = img.height
        this.setUpdates({
          items: [{
            updateAction: 'set',
            index: itemIndex,
            ...this.getItemData(itemIndex)
          }],
          files: [{
            updateAction: justUploaded ? 'add' : 'set',
            index: this.getTotalFileIndex(itemIndex, fileIndex),
            ...this.getFileData(itemIndex, fileIndex)
          }]
        })
      }
      const reader = new FileReader()
      reader.onload = e => {
        // @ts-ignore
        fileObj.src = e.target.result
        img.src = fileObj.src
      }
      reader.readAsDataURL(fileObj.file)
    },
    updateVideoSrc(itemIndex, fileIndex, justUploaded) {
      const fileObj = this.items[itemIndex].files[fileIndex]
      if (!fileObj.file) return
      const reader = new FileReader()
      reader.onload = e => {
        // @ts-ignore
        fileObj.src = e.target.result
        this.$forceUpdate()
        this.$nextTick(() => {
          const videotag = this.$refs[`video-source-${itemIndex}-${fileIndex}`][0]
          videotag.load()
          videotag.oncanplay = () => {
            // @ts-ignore
            fileObj.height = videotag.videoHeight
            // @ts-ignore
            fileObj.width = videotag.videoWidth
            this.setUpdates({
              items: [{
                updateAction: 'set',
                index: itemIndex,
                ...this.getItemData(itemIndex)
              }],
              files: [{
                updateAction: justUploaded ? 'add' : 'set',
                index: this.getTotalFileIndex(itemIndex, fileIndex),
                ...this.getFileData(itemIndex, fileIndex)
              }]
            })
          }
        })
      }
      reader.readAsDataURL(fileObj.file)
    },
    updateEditorContent(index) {
      if (this.$refs[`editor-${index}`] && this.$refs[`editor-${index}`].length > 0) {
        this.$refs[`editor-${index}`][0]._data.editor.setContent(this.editorContent[index])
      }
    },
    finishedDragging(evt, doUpdate) {
      if (evt.oldIndex === evt.newIndex) {
        return
      }
      const oldTextLocations = Object.keys(this.editorContent).map(elem => Number(elem)).sort()
      const newTextLocations = [...oldTextLocations]
      let indexOld
      const movingText = this.items[evt.newIndex].type === itemTypes[3].id
      if (movingText) {
        indexOld = oldTextLocations.indexOf(evt.oldIndex)
        if (indexOld < 0) {
          return
        }
        newTextLocations.splice(indexOld, 1)
      }
      let inserted = false
      for (let i = 0; i < newTextLocations.length; i++) {
        let addAbove
        if (newTextLocations[i] < evt.oldIndex && newTextLocations[i] >= evt.newIndex) {
          newTextLocations[i]++
          addAbove = false
        } else if (newTextLocations[i] > evt.oldIndex && newTextLocations[i] <= evt.newIndex) {
          newTextLocations[i]--
          addAbove = true
        }
        if (newTextLocations[i] === evt.newIndex && movingText) {
          newTextLocations.splice(i + (addAbove ? 1 : -1), 0, evt.newIndex)
          i++
          inserted = true
        }
      }
      if (!inserted && movingText) {
        newTextLocations.splice(indexOld, 0, evt.newIndex)
      }
      const newEditorContent = {}
      for (let i = 0; i < newTextLocations.length; i++) {
        newEditorContent[newTextLocations[i]] = this.editorContent[oldTextLocations[i]]
      }
      this.editorContent = newEditorContent
      for (let i = 0; i < newTextLocations.length; i++) {
        this.updateEditorContent(newTextLocations[i])
      }
      if (doUpdate) {
        this.setUpdates({
          items: [{
            updateAction: 'move',
            index: evt.oldIndex,
            newIndex: evt.newIndex
          }]
        })
        this.focusIndex = evt.newIndex
      } else if (this.focusIndex === evt.oldIndex) {
        this.focusIndex = evt.newIndex
      } else if (evt.newIndex >= this.focusIndex && evt.oldIndex <= this.focusIndex) {
        this.focusIndex--
      } else if (evt.newIndex <= this.focusIndex && evt.oldIndex >= this.focusIndex) {
        this.focusIndex++
      }
    },
    focusItem(evt, itemIndex) {
      this.focusIndex = itemIndex
    },
    getItemTypeLabel(itemIndex) {
      const itemType = itemTypes.find((elem) => elem.id === this.items[itemIndex].type)
      if (itemType) {
        return itemType.label
      }
      return 'Unknown Type'
    },
    selectItemType(evt, itemIndex, type) {
      evt.preventDefault()
      const oldType = this.items[itemIndex].type
      if ((type.id === itemTypes[0].id ||
          type.id === itemTypes[1].id)) {
        if (!(oldType === itemTypes[0].id ||
          oldType === itemTypes[1].id)) {
          this.items[itemIndex].options = ['']
        }
      } else {
        this.items[itemIndex].options = []
      }
      if (type.id === itemTypes[3].id) {
        this.items[itemIndex].type = type.id
        this.$forceUpdate()
        this.$nextTick(() => {
          if (!this.editorContent.hasOwnProperty(itemIndex)) {
            this.editorContent[itemIndex] = this.$refs[`editor-${itemIndex}`][0]._data.editor.getHTML()
          }
        })
      } else {
        this.deleteEditorData(itemIndex)
        this.items[itemIndex].text = ''
        this.items[itemIndex].type = type.id
      }
      if (type.id === itemTypes[7].id) {
        if (this.validDisplayFiles.includes(this.items[itemIndex].files[0].type)) {
          this.updateFileSrc(itemIndex, 0, false)
        } else {
          const fileObj = this.items[itemIndex].files[0]
          if (fileObj.uploaded) {
            this.deleteFile(itemIndex, 0)
          }
          this.items[itemIndex].files = [clone(defaultFile)]
        }
      } else if (this.items[itemIndex].files[0].uploaded &&
        type.id !== itemTypes[6].id) {
        // no file uploads
        this.deleteFile(itemIndex, 0)
      }
      if (type.id !== itemTypes[3].id) {
        this.$forceUpdate()
      }
      this.setUpdates({
        items: [{
          updateAction: 'set',
          index: itemIndex,
          ...this.getItemData(itemIndex)
        }]
      })
    },
    deleteEditorData(itemIndex) {
      if (this.editorContent.hasOwnProperty(itemIndex)) {
        delete this.editorContent[itemIndex]
      }
    },
    updatedText(newText, itemIndex) {
      this.editorContent[itemIndex] = newText
      this.setUpdates({
        items: [{
          updateAction: 'set',
          index: itemIndex,
          ...this.getItemData(itemIndex)
        }]
      })
    },
    addItem(evt, doUpdate) {
      if (evt) {
        evt.preventDefault()
      }
      const newItem = clone(defaultItem)
      if (newItem.type === itemTypes[0].id ||
          newItem.type === itemTypes[1].id) {
        newItem.options.push('')
      }
      this.items.push(newItem)
      this.focusIndex = this.items.length - 1
      if (doUpdate) {
        const newItemUpdateObj = clone(newItem)
        newItemUpdateObj.updateAction = 'add'
        newItemUpdateObj.files = []
        this.setUpdates({
          items: [newItemUpdateObj]
        })
      }
    },
    removeItem(evt, itemIndex, doUpdate) {
      if (evt) {
        evt.preventDefault()
      }
      if (this.items[itemIndex].type === itemTypes[3].id) {
        this.deleteEditorData(itemIndex)
      }
      if ((this.items[itemIndex].type === itemTypes[5].id ||
          this.items[itemIndex].type === itemTypes[6].id ||
          this.items[itemIndex].type === itemTypes[7].id) &&
          this.items[itemIndex].files[0].uploaded) {
        if (doUpdate) {
          this.deleteFile(itemIndex, 0)
        } else {
          this.items[itemIndex].files[0].src = null
          this.items[itemIndex].files[0].uploaded = false
        }
      }
      if (doUpdate) {
        this.setUpdates({
          items: [{
            updateAction: 'remove',
            index: itemIndex
          }]
        })
      }
      this.items.splice(itemIndex, 1)
      if (this.focusIndex >= this.items.length) {
        if (this.items.length > 0) {
          this.focusIndex = this.items.length - 1
        } else {
          this.focusIndex = 0
        }
      }
    },
    addOption(evt, itemIndex) {
      evt.preventDefault()
      this.items[itemIndex].options.push('')
      this.setUpdates({
        items: [{
          updateAction: 'set',
          index: itemIndex,
          ...this.getItemData(itemIndex)
        }]
      })
      this.$forceUpdate()
    },
    removeOption(evt, itemIndex, optionIndex) {
      evt.preventDefault()
      this.items[itemIndex].options.splice(optionIndex, 1)
      this.setUpdates({
        items: [{
          updateAction: 'set',
          index: itemIndex,
          ...this.getItemData(itemIndex)
        }]
      })
      this.$forceUpdate()
    }
  }
})
</script>

<style lang="scss">
.sampleimage {
  max-width: 30rem;
}
.drag-area {
  padding: 5px;
  cursor: move;
  .icon-grip {
    color: lightgray;
    font-size: 20px;
    margin-left: 37px;
  }
}
.ghost {
  opacity: 0;
}
.separate-items {
  margin-top: 2rem;
  margin-bottom: 2rem;
}
.disable-button {
  pointer-events: none;
}
.add-button {
  margin-right: 5rem;
  height: 3rem;
  width: 3rem;
  text-align: center;
  line-height: 50%;
}
.send-button {
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
