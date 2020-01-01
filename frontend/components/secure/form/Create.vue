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
              @end="finishedDragging"
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
                                @input="updateFileSrc(index, 0)"
                                placeholder="Choose a file or drop it here..."
                                drop-placeholder="Drop file here..."
                                style="max-width:30rem;"
                                class="mb-2"
                              />
                            </b-col>
                          </b-row>
                          <b-row>
                            <b-col>
                              <a
                                v-if="item.files[0].file"
                                :href="getFileURL(index)"
                                :download="items[index].files[0].name"
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
                                @input="updateFileSrc(index, 0)"
                                placeholder="Choose an image or drop it here..."
                                drop-placeholder="Drop image here..."
                                style="max-width:30rem;"
                                class="mb-2"
                              />
                            </b-col>
                          </b-row>
                          <b-row>
                            <b-col style="padding-left:0;">
                              <b-img
                                v-if="
                                  item.files[0].file &&
                                    item.files[0].src &&
                                    checkImageType(item.files[0].type)
                                "
                                :src="item.files[0].src"
                                class="mt-2 mb-2 sampleimage"
                              />
                              <video
                                v-else-if="
                                  item.files[0].src &&
                                    item.files[0].id &&
                                    item.files[0].type &&
                                    checkVideoType(item.files[0].type)
                                "
                                :ref="`video-source-${index}-${0}`"
                                :type="blog.files[index].type"
                                :src="blog.files[index].src"
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
                            @click="(evt) => removeItem(evt, index)"
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
                    @click="addItem"
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
import clonedeep from 'lodash.clonedeep'
import gql from 'graphql-tag'
import TextEditor from '~/components/secure/form/TextEditor.vue'
import PageLoading from '~/components/PageLoading.vue'
import { defaultItemName, validfiles, validimages, validDisplayFiles } from '~/assets/config'

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
    id: 'image',
    label: 'Image'
  },
]
const defaultFile = {
  id: '',
  name: '',
  width: null,
  height: null,
  type: '',
  file: null,
  src: null
}
const clone = (obj) => {
  const newObj = clonedeep(obj)
  delete newObj.__typename
  return newObj
}
const defaultItem = {
  question: '',
  type: itemTypes[0].id,
  options: [],
  text: [],
  required: false,
  files: [clone(defaultFile)]
}
export default Vue.extend({
  name: 'Create',
  components: {
    TextEditor,
    PageLoading
  },
  props: {
    getInitialData: {
      type: Boolean,
      default: true
    },
    projectId: {
      type: String,
      default: null
    },
    formId: {
      type: String,
      default: null
    }
  },
  data() {
    return {
      validDisplayFiles,
      validfiles,
      validimages,
      loading: true,
      itemTypes,
      focusIndex: 0,
      editorContent: {},
      name: '',
      items: [],
      multiple: false
    }
  },
  mounted() {
    if (this.getInitialData) {
      this.$apollo.query({query: gql`
        query form($id: String!){form(id: $id){name,items{question,type,options,text,required,files},multiple,files{id,name,width,height,type}} }
        `, variables: {id: this.formId}})
        .then(({ data }) => {
          this.name = data.form.name
          const newEditorContent = {}
          for (let i = 0; i < data.form.items.length; i++) {
            if (data.form.items[i].type === itemTypes[3].id) {
              newEditorContent[i] = data.form.items[i].text
            }
            data.form.items[i].files = [clone(defaultFile)]
          }
          this.editorContent = newEditorContent
          this.items = data.form.items
          this.multiple = data.form.multiple
          this.loading = false
          this.$nextTick(() => {
            const newTextLocations = Object.keys(this.editorContent)
            for (let i = 0; i < newTextLocations.length; i++) {
              this.$refs[`editor-${newTextLocations[i]}`][0]._data.editor.setContent(this.editorContent[newTextLocations[i]])
            }
          })
        }).catch(err => {
          console.log(err.message)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    } else {
      this.name = defaultItemName
      this.loading = false
    }
  },
  methods: {
    submit(evt) {
      evt.preventDefault()
      for (let i = 0; i < this.items.length; i++) {
        if (this.items[i].type === itemTypes[3].id) {
          this.items[i].text = this.editorContent[i]
        }
        delete this.items[i].__typename
        this.items[i].files = []
      }
      // send images first
      // then send json object
      // console.log(this.items)
      const files = []
      this.$apollo.mutate({mutation: gql`
        mutation updateForm($id: String!, $name: String!, $items: [ItemInput!]!, $multiple: Boolean!, $files: [FileInput!]!)
        {updateForm(id: $id, name: $name, items: $items, multiple: $multiple, files: $files){id} }
        `, variables: {id: this.formId, name: this.name, items: this.items, multiple: this.multiple, files}})
        .then(({ data }) => {
          console.log('updated!')
        }).catch(err => {
          console.error(err)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    },
    getFileURL(itemIndex) {
      return ''
      /*
      if (this.items[itemIndex].files[0].id.lenquerygth > 0) {
        // it's uploaded. create link to get from cloud
        return ''
      }
      return URL.createObjectURL(this.items[itemIndex].files[0].file)
      */
    },
    checkImageType(type) {
      return /^image\/.*$/.test(type)
    },
    checkVideoType(type) {
      return /^video\/.*$/.test(type)
    },
    updateFileSrc(itemIndex, fileIndex) {
      const fileObj = this.items[itemIndex].files[fileIndex]
      if (fileObj.file && !fileObj.src) {
        fileObj.type = fileObj.file.type
        if (this.checkVideoType(fileObj.type))
          this.updateVideoSrc(itemIndex, fileIndex)
        else if (this.checkImageType(fileObj.type))
          this.updateImageSrc(itemIndex, fileIndex)
      }
    },
    updateImageSrc(itemIndex, fileIndex) {
      const fileObj = this.items[itemIndex].files[fileIndex]
      if (!fileObj.file) return
      const img = new Image()
      img.onload = () => {
        console.log('image loaded')
        fileObj.width = img.width
        fileObj.height = img.height
        console.log(`image width: ${fileObj.width}, height: ${fileObj.height}`)
      }
      const reader = new FileReader()
      reader.onload = e => {
        // @ts-ignore
        fileObj.src = e.target.result
        img.src = fileObj.src
      }
      reader.readAsDataURL(fileObj.file)
      console.log('done')
    },
    updateVideoSrc(itemIndex, fileIndex) {
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
          }
        })
      }
      reader.readAsDataURL(fileObj.file)
      console.log('done')
    },
    finishedDragging(evt) {
      if (evt.oldIndex === evt.newIndex) {
        return;
      }
      const oldTextLocations = Object.keys(this.editorContent).map(elem => Number(elem)).sort()
      const newTextLocations = [...oldTextLocations]
      let indexOld
      const movingText = this.items[evt.newIndex].type === itemTypes[3].id
      if (movingText) {
        indexOld = oldTextLocations.indexOf(evt.oldIndex)
        if (indexOld < 0) {
          return;
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
      for (let i = 0; i < newTextLocations.length; i++) {
        this.$refs[`editor-${newTextLocations[i]}`][0]._data.editor.setContent(newEditorContent[newTextLocations[i]])
      }
      this.editorContent = newEditorContent
      this.focusIndex = evt.newIndex
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
      if (type.id === itemTypes[3].id) {
        this.items[itemIndex].options = []
        this.items[itemIndex].type = type.id
        this.$nextTick(() => {
          if (!this.editorContent.hasOwnProperty(itemIndex)) {
            this.editorContent[itemIndex] = this.$refs[`editor-${itemIndex}`][0]._data.editor.getHTML()
          }
        })
      } else {
        this.deleteEditorData(itemIndex)
        this.items[itemIndex].text = ''
        this.items[itemIndex].options = ['']
        this.items[itemIndex].type = type.id
        if (type.id === itemTypes[7].id) {
          if (this.validDisplayFiles.includes(this.items[itemIndex].files[0].type)) {
            this.updateFileSrc(itemIndex, 0)
          } else {
            this.items[itemIndex].files = [clone(defaultFile)]
          }
        }
      }
    },
    deleteEditorData(itemIndex) {
      if (this.editorContent.hasOwnProperty(itemIndex)) {
        delete this.editorContent[itemIndex]
      }
    },
    updatedText(newText, itemIndex) {
      this.editorContent[itemIndex] = newText
    },
    addItem(evt) {
      evt.preventDefault()
      const newItem = clone(defaultItem)
      if (newItem.type !== itemTypes[3].id) {
        newItem.options.push('')
      }
      this.items.push(newItem)
      this.focusIndex = this.items.length - 1
    },
    removeItem(evt, itemIndex) {
      evt.preventDefault()
      if (this.items[itemIndex].type === itemTypes[3].id) {
        this.deleteEditorData(itemIndex)
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
    },
    removeOption(evt, itemIndex, optionIndex) {
      evt.preventDefault()
      this.items[itemIndex].options.splice(optionIndex, 1)
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
