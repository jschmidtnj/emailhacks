<template>
  <div class="container-fluid">
    <div id="admin-cards" class="container">
      <div class="row my-4">
        <div class="col-lg-6 my-2">
          <section class="card h-100 py-0">
            <div class="card-body">
              <b-form @submit="manageblogs" @reset="resetblogs">
                <span class="card-text">
                  <h2 class="mb-4">{{ mode }} Blog</h2>
                  <b-form-group>
                    <label>Content</label>
                    <span>
                      <b-form-textarea
                        v-model="blog.content"
                        :state="!$v.blog.content.$invalid"
                        type="text"
                        class="form-control"
                        aria-describedby="contentfeedback"
                        placeholder="Enter content..."
                        rows="5"
                        max-rows="15"
                      />
                    </span>
                    <b-form-invalid-feedback
                      id="contentfeedback"
                      :state="!$v.blog.content.$invalid"
                    >
                      <div v-if="!$v.blog.content.required">
                        content is required
                      </div>
                      <div v-else-if="!$v.blog.content.minLength">
                        content must have at least
                        {{ $v.blog.content.$params.minLength.min }} characters
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-form-group>
                    <label class="form-required">Author</label>
                    <span>
                      <b-form-input
                        id="author"
                        v-model="blog.author"
                        :state="!$v.blog.author.$invalid"
                        type="text"
                        class="form-control"
                        aria-describedby="authorfeedback"
                        placeholder="author"
                      />
                    </span>
                    <b-form-invalid-feedback
                      id="authorfeedback"
                      :state="!$v.blog.author.$invalid"
                    >
                      <div v-if="!$v.blog.author.required">
                        author is required
                      </div>
                      <div v-else-if="!$v.blog.author.minLength">
                        author must have at least
                        {{ $v.blog.author.$params.minLength.min }} characters
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-form-group>
                    <label class="form-required">Title</label>
                    <span>
                      <b-form-input
                        id="title"
                        v-model="blog.title"
                        :state="!$v.blog.title.$invalid"
                        type="text"
                        class="form-control"
                        aria-describedby="titlefeedback"
                        placeholder="title"
                      />
                    </span>
                    <b-form-invalid-feedback
                      id="titlefeedback"
                      :state="!$v.blog.title.$invalid"
                    >
                      <div v-if="!$v.blog.title.required">
                        title is required
                      </div>
                      <div v-else-if="!$v.blog.title.minLength">
                        title must have at least
                        {{ $v.blog.title.$params.minLength.min }} characters
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-form-group>
                    <label class="form-required">Caption</label>
                    <span>
                      <b-form-input
                        id="caption"
                        v-model="blog.caption"
                        :state="!$v.blog.caption.$invalid"
                        type="text"
                        class="form-control"
                        aria-describedby="captionfeedback"
                        placeholder="caption"
                      />
                    </span>
                    <b-form-invalid-feedback
                      id="captionfeedback"
                      :state="!$v.blog.caption.$invalid"
                    >
                      <div v-if="!$v.blog.caption.required">
                        caption is required
                      </div>
                      <div v-else-if="!$v.blog.caption.minLength">
                        caption must have at least
                        {{ $v.blog.caption.$params.minLength.min }} characters
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-form-group>
                    <label class="form-required">Theme Color</label>
                    <span>
                      <client-only>
                        <color-picker
                          v-model="blog.color"
                          aria-describedby="colorfeedback"
                        />
                      </client-only>
                    </span>
                    <b-form-invalid-feedback
                      id="colorfeedback"
                      :state="!$v.blog.color.$invalid"
                    >
                      <div v-if="!$v.blog.color.required">
                        color is required
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-form-group>
                    <label class="form-required">Categories</label>
                    <span>
                      <client-only>
                        <v-select
                          v-model="blog.categories"
                          :options="categoryOptions"
                          :multiple="true"
                          :taggable="true"
                          aria-describedby="categoryfeedback"
                        />
                      </client-only>
                    </span>
                    <b-form-invalid-feedback
                      id="categoryfeedback"
                      :state="!$v.blog.categories.$invalid"
                    >
                      <div v-if="!$v.blog.categories.required">
                        categories is required
                      </div>
                      <div v-else-if="!$v.blog.categories.minLength">
                        categories must have at least
                        {{ $v.blog.categories.$params.$each.minLength.min }}
                        characters
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-form-group>
                    <label class="form-required">Tags</label>
                    <span>
                      <client-only>
                        <v-select
                          v-model="blog.tags"
                          :options="tagOptions"
                          :multiple="true"
                          :taggable="true"
                          aria-describedby="tagfeedback"
                        />
                      </client-only>
                    </span>
                    <b-form-invalid-feedback
                      id="tagfeedback"
                      :state="!$v.blog.tags.$invalid"
                    >
                      <div v-if="!$v.blog.tags.required">tags is required</div>
                      <div v-else-if="!$v.blog.tags.minLength">
                        tags must have at least
                        {{ $v.blog.tags.$params.$each.minLength.min }}
                        characters
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-img
                    v-if="blog.heroimage.file && blog.heroimage.src"
                    :src="blog.heroimage.src"
                    class="sampleimage"
                  />
                  <b-form-group>
                    <label class="form-required">Hero Image</label>
                    <span>
                      <b-form-file
                        v-model="blog.heroimage.file"
                        :accept="validimages.join(', ')"
                        :state="!$v.blog.heroimage.$invalid"
                        @input="
                          blog.heroimage.uploaded = false
                          updateFileSrc(blog.heroimage)
                        "
                        class="mb-2 form-control"
                        aria-describedby="heroimagefeedback"
                        placeholder="Choose an image..."
                        drop-placeholder="Drop image here..."
                      />
                    </span>
                    <b-form-invalid-feedback
                      id="heroimagefeedback"
                      :state="!$v.blog.heroimage.$invalid"
                    >
                      <div v-if="!$v.blog.heroimage.required">
                        hero image is required
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-img
                    v-if="blog.tileimage.file && blog.tileimage.src"
                    :src="blog.tileimage.src"
                    class="sampleimage"
                  />
                  <b-form-group>
                    <label class="form-required">Tile Image</label>
                    <span>
                      <b-form-file
                        v-model="blog.tileimage.file"
                        :accept="validimages.join(', ')"
                        :state="!$v.blog.tileimage.$invalid"
                        @input="
                          blog.tileimage.uploaded = false
                          updateFileSrc(blog.tileimage)
                        "
                        class="mb-2 form-control"
                        aria-describedby="tileimagefeedback"
                        placeholder="Choose an image..."
                        drop-placeholder="Drop image here..."
                      />
                    </span>
                    <b-form-invalid-feedback
                      id="tileimagefeedback"
                      :state="!$v.blog.tileimage.$invalid"
                    >
                      <div v-if="!$v.blog.tileimage.required">
                        tile image is required
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <h4 class="mt-4">Files</h4>
                  <div
                    v-for="(filevalue, index) in $v.blog.files.$each.$iter"
                    :key="`file-${index}`"
                  >
                    <b-img
                      v-if="
                        blog.files[index].src &&
                          blog.files[index].type &&
                          checkImageType(blog.files[index].type)
                      "
                      :src="blog.files[index].src"
                      class="sampleimage"
                    />
                    <video
                      v-else-if="
                        blog.files[index].src &&
                          blog.files[index].id &&
                          blog.files[index].type &&
                          checkVideoType(blog.files[index].type)
                      "
                      :ref="`video-source-${blog.files[index].id}`"
                      :type="blog.files[index].type"
                      :src="blog.files[index].src"
                      controls
                      autoplay
                      class="sampleimage"
                      allowfullscreen
                    />
                    <br />
                    <code
                      v-if="
                        blog.files[index].type === 'image/gif' &&
                          (blog.files[index].file ||
                            blog.files[index].uploaded) &&
                          blog.files[index].name &&
                          blog.files[index].width &&
                          blog.files[index].height &&
                          blog.files[index].id
                      "
                      >{{ getGifTag(blog.files[index]) }}</code
                    >
                    <code
                      v-else-if="
                        blog.files[index].file &&
                          blog.files[index].type &&
                          checkImageType(blog.files[index].type) &&
                          blog.files[index].name &&
                          blog.files[index].width &&
                          blog.files[index].height &&
                          blog.files[index].id
                      "
                      >{{ getImageTag(blog.files[index]) }}</code
                    >
                    <code
                      v-else-if="
                        (blog.files[index].file ||
                          blog.files[index].uploaded) &&
                          blog.files[index].type &&
                          checkVideoType(blog.files[index].type) &&
                          blog.files[index].name &&
                          blog.files[index].width &&
                          blog.files[index].height &&
                          blog.files[index].id &&
                          blog.files[index].type
                      "
                      >{{ getVideoTag(blog.files[index]) }}</code
                    >
                    <code
                      v-else-if="
                        (blog.files[index].file ||
                          blog.files[index].uploaded) &&
                          blog.files[index].name &&
                          blog.files[index].id
                      "
                      >{{ getFileTag(blog.files[index]) }}</code
                    >
                    <b-form-group class="mb-2">
                      <label class="form-required">File Name</label>
                      <span>
                        <b-form-input
                          v-model="blog.files[index].name"
                          :state="!filevalue.name.$invalid"
                          @input="blog.files[index].uploaded = false"
                          type="text"
                          class="form-control"
                          placeholder="name"
                        />
                      </span>
                      <b-form-invalid-feedback
                        :state="!filevalue.name.$invalid"
                      >
                        <div v-if="!filevalue.name.required">
                          file name is required
                        </div>
                        <div v-else-if="!filevalue.name.minLength">
                          file name must have at least
                          {{ filevalue.name.$params.minLength.min }} characters
                        </div>
                      </b-form-invalid-feedback>
                    </b-form-group>
                    <b-form-group>
                      <label class="form-required">File</label>
                      <span>
                        <b-form-file
                          v-model="blog.files[index].file"
                          :accept="validfiles.join(', ')"
                          :state="!filevalue.file.$invalid"
                          @input="
                            blog.files[index].uploaded = false
                            updateFileSrc(blog.files[index])
                          "
                          class="mb-2 form-control"
                          placeholder="Choose a file..."
                          drop-placeholder="Drop file here..."
                        />
                      </span>
                      <b-form-invalid-feedback
                        :state="!filevalue.file.$invalid"
                      >
                        <div v-if="!filevalue.file.gotFile">
                          file is required
                        </div>
                      </b-form-invalid-feedback>
                    </b-form-group>
                  </div>
                  <b-container class="mt-4">
                    <b-row>
                      <b-col>
                        <b-btn
                          @click="
                            blog.files.push({
                              name: '',
                              file: null,
                              uploaded: false,
                              id: createId(),
                              src: null,
                              width: null,
                              height: null,
                              type: null
                            })
                          "
                          variant="primary"
                          class="mr-2"
                        >
                          <client-only>
                            <font-awesome-icon
                              class="mr-2 arrow-size-edit"
                              icon="plus-circle"
                            /> </client-only
                          >Add
                        </b-btn>
                        <b-btn
                          :disabled="blog.files.length === 0"
                          @click="removeFile"
                          variant="primary"
                          class="mr-2"
                        >
                          <client-only>
                            <font-awesome-icon
                              class="mr-2 arrow-size-edit"
                              icon="times"
                            /> </client-only
                          >Remove
                        </b-btn>
                      </b-col>
                    </b-row>
                  </b-container>
                  <b-container class="mt-4">
                    <b-row>
                      <b-col>
                        <b-btn
                          :disabled="$v.blog.$invalid || submitting"
                          variant="primary"
                          type="submit"
                        >
                          <client-only>
                            <font-awesome-icon
                              class="mr-2 arrow-size-edit"
                              icon="angle-double-right"
                            /> </client-only
                          >Submit
                        </b-btn>
                        <b-btn variant="primary" type="reset" class="mr-4">
                          <client-only>
                            <font-awesome-icon
                              class="mr-2 arrow-size-edit"
                              icon="times"
                            /> </client-only
                          >Clear
                        </b-btn>
                      </b-col>
                    </b-row>
                  </b-container>
                </span>
              </b-form>
            </div>
          </section>
        </div>
        <div class="col-lg-6 my-2">
          <section class="card h-100 py-0">
            <div class="card-body">
              <b-form @submit="searchblogs" @reset="clearsearch">
                <span class="card-text">
                  <div
                    id="content-rendered"
                    v-if="blog.content !== ''"
                    class="mb-4"
                  >
                    <h2 class="mb-4">Content</h2>
                    <vue-markdown
                      :source="blog.content"
                      @rendered="updateMarkdown"
                      class="mb-4 markdown"
                    />
                  </div>
                  <h2 class="mb-4">Search</h2>
                  <b-form-group>
                    <label class="form-required">Query</label>
                    <span>
                      <b-form-input
                        v-model="search"
                        :state="!$v.search.$invalid"
                        type="text"
                        class="form-control mb-2"
                        aria-describedby="searchfeedback"
                        placeholder="search..."
                      />
                    </span>
                    <b-form-invalid-feedback
                      id="searchfeedback"
                      :state="!$v.search.$invalid"
                    >
                      <div v-if="!$v.search.required">query is required</div>
                      <div v-else-if="!$v.search.minLength">
                        query must have at least
                        {{ $v.search.$params.minLength.min }} characters
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-btn
                    :disabled="$v.search.$invalid"
                    variant="primary"
                    type="submit"
                    class="mr-4"
                  >
                    <client-only>
                      <font-awesome-icon
                        class="mr-2 arrow-size-edit"
                        icon="angle-double-right"
                      /> </client-only
                    >Search
                  </b-btn>
                  <b-btn variant="primary" type="reset" class="mr-4">
                    <client-only>
                      <font-awesome-icon
                        class="mr-2 arrow-size-edit"
                        icon="times"
                      /> </client-only
                    >Clear
                  </b-btn>
                  <br />
                  <br />
                </span>
              </b-form>
              <b-table
                :items="searchresults"
                :fields="fields"
                :current-page="currentpage"
                :per-page="numperpage"
                show-empty
                stacked="md"
              >
                <template v-slot:cell(name)="data">
                  {{ data.value }}
                </template>
                <template v-slot:cell(created)="data">
                  {{ formatDate(data.value, 'M/d/yyyy') }}
                </template>
                <template v-slot:cell(id)="data">
                  <nuxt-link
                    :to="`/blog/${data.value}`"
                    class="btn btn-primary btn-sm no-underline"
                  >
                    {{ row.value }}
                  </nuxt-link>
                </template>
                <template v-slot:cell(actions)="data">
                  <b-button @click="editBlog(data.item)" size="sm" class="mr-1">
                    Edit
                  </b-button>
                  <b-button @click="deleteBlog(data.item)" size="sm">
                    Del
                  </b-button>
                </template>
              </b-table>
              <b-row class="mb-2">
                <b-col md="6" class="my-1">
                  <b-pagination
                    v-model="currentpage"
                    :total-rows="searchresults.length"
                    :per-page="numperpage"
                    class="my-0"
                  />
                </b-col>
              </b-row>
            </div>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import { validationMixin } from 'vuelidate'
import { required, minLength } from 'vuelidate/lib/validators'
import VueMarkdown from 'vue-markdown'
import Prism from 'prismjs'
import { format } from 'date-fns'
import uuid from 'uuid/v1'
import axios from 'axios'
import { Chrome } from 'vue-color'
import LazyLoad from 'vanilla-lazyload'
import { ObjectID } from 'bson'
import {
  cloudStorageURLs,
  options,
  defaultColor,
  staticstorageindexes,
  validimages,
  validfiles,
  paths
} from '~/assets/config'
const gotFile = (_, vm) => vm.uploaded || vm.src !== null
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
const lazyLoadInstance = new LazyLoad({
  elements_selector: '.lazy'
})
/**
 * blogs edit
 */
const modetypes = {
  add: 'Add',
  edit: 'Edit',
  delete: 'Delete'
}
const originalHero = {
  name: 'hero',
  uploaded: false,
  file: null,
  id: uuid(),
  src: null,
  width: null,
  height: null
}
const originalTile = Object.assign({}, originalHero)
originalTile.id = uuid()
originalTile.name = 'tile'
export default Vue.extend({
  name: 'BlogsEdit',
  // @ts-ignore
  layout: 'admin',
  components: {
    VueMarkdown,
    'color-picker': Chrome
  },
  mixins: [validationMixin],
  // @ts-ignore
  data() {
    return {
      submitting: false,
      modetypes,
      mode: modetypes.add,
      blogid: new ObjectID().toString(),
      search: '',
      type: 'blog',
      searchresults: [],
      currentpage: 1,
      numperpage: 10,
      categoryOptions: options.categoryOptions,
      tagOptions: options.tagOptions,
      validimages,
      validfiles,
      paths,
      fields: [
        {
          key: 'title',
          label: 'Title',
          sortable: true
        },
        {
          key: 'created',
          label: 'Created',
          sortable: true
        },
        {
          key: 'id',
          label: 'ID',
          sortable: true
        },
        {
          key: 'actions',
          label: 'Actions',
          sortable: false
        }
      ],
      blog: {
        title: '',
        content: '',
        caption: '',
        color: defaultColor,
        author: '',
        tags: [],
        categories: [],
        heroimage: Object.assign({}, originalHero),
        tileimage: Object.assign({}, originalTile),
        files: []
      }
    }
  },
  // @ts-ignore
  validations: {
    search: {
      required,
      minLength: minLength(3)
    },
    blog: {
      title: {
        required,
        minLength: minLength(3)
      },
      author: {
        required,
        minLength: minLength(3)
      },
      caption: {
        required,
        minLength: minLength(3)
      },
      content: {
        required,
        minLength: minLength(10)
      },
      color: {
        required
      },
      heroimage: {
        file: {}
      },
      tileimage: {
        file: {
          required
        }
      },
      tags: {
        $each: {
          required
        }
      },
      categories: {
        $each: {
          required
        }
      },
      files: {
        $each: {
          name: {
            required,
            minLength: minLength(3)
          },
          file: {
            gotFile
          }
        }
      }
    }
  },
  // @ts-ignore
  head() {
    const title = 'Admin Edit Blog'
    const description = 'admin page for editing Blogs'
    const image = `${seo.url}/icon.png`
    return {
      title,
      meta: [
        { property: 'og:title', content: title },
        { property: 'og:description', content: description },
        {
          property: 'og:image',
          content: image
        },
        { name: 'twitter:title', content: title },
        {
          name: 'twitter:description',
          content: description
        },
        {
          name: 'twitter:image',
          content: image
        },
        { hid: 'description', name: 'description', content: description }
      ]
    }
  },
  /* eslint-disable */
  methods: {
    checkImageType(type) {
      return /^image\/.*$/.test(type)
    },
    checkVideoType(type) {
      return /^video\/.*$/.test(type)
    },
    updateMarkdown() {
      this.$nextTick(() => {
        Prism.highlightAll()
        if (lazyLoadInstance) {
          console.log('update lazyload')
          lazyLoadInstance.update()
        }
      })
    },
    createId() {
      return uuid()
    },
    mongoidToDate(id) {
      return parseInt(id.substring(0, 8), 16) * 1000
    },
    formatDate(dateUTC, formatStr) {
      return format(dateUTC, formatStr)
    },
    getImageTag(image) {
      return `<img data-src="${cloudStorageURLs.blogs}/${
        staticstorageindexes.blogfiles
      }/${this.blogid}/${image.id + this.paths.original}" src="${
        cloudStorageURLs.blogs
      }/${staticstorageindexes.blogfiles}/${this.blogid}/${
        image.id + this.paths.blur}" class="lazy img-fluid" alt="${
        image.name
      }" data-width="${image.width}" data-height="${image.height}">`
    },
    getGifTag(gif) {
      return `<img data-src="${cloudStorageURLs.blogs}/${
        staticstorageindexes.blogfiles
      }/${this.blogid}/${gif.id + this.paths.original}" placeholder-original="${
        cloudStorageURLs.blogs
      }/${staticstorageindexes.blogfiles}/${this.blogid}/${
        gif.id + this.paths.placeholder + this.paths.original}" src="${
        cloudStorageURLs.blogs
      }/${staticstorageindexes.blogfiles}/${this.blogid}/${
        gif.id + this.paths.placeholder + this.paths.blur}" class="lazy img-fluid gif" alt="${
        gif.name
      }" data-width="${gif.width}" data-height="${gif.height}">`
    },
    getVideoTag(video) {
      return `<video class="img-fluid" data-width="${video.width}" data-height="${
        video.height
      }" alt="${video.name}" controls allowfullscreen><source src="${cloudStorageURLs.blogs}/${
        staticstorageindexes.blogfiles
      }/${this.blogid}/${video.id}#t=0.1" type="${video.type}" /></video>`
    },
    getFileTag(file) {
      return `<a href="${cloudStorageURLs.blogs}/${
        staticstorageindexes.blogfiles
      }/${this.blogid}/${file.id}" target="_blank">download</a>`
    },
    updateFileSrc(file) {
      if (file.file) {
        file.type = file.file.type
        if (this.checkVideoType(file.type))
          this.updateVideoSrc(file)
        else
          this.updateImageSrc(file) 
      }
    },
    updateImageSrc(image) {
      console.log('start image src')
      if (!image.file) return
      const img = new Image()
      img.onload = () => {
        console.log('image loaded')
        image.width = img.width
        image.height = img.height
        console.log(`image width: ${image.width}, height: ${image.height}`)
      }
      const reader = new FileReader()
      reader.onload = e => {
        // @ts-ignore
        image.src = e.target.result
        img.src = image.src
      }
      reader.readAsDataURL(image.file)
      console.log('done')
    },
    updateVideoSrc(video) {
      if (!video.file) return
      const reader = new FileReader()
      reader.onload = e => {
        // @ts-ignore
        video.src = e.target.result
        this.$forceUpdate()
        this.$nextTick(() => {
          const videotag = this.$refs[`video-source-${video.id}`][0]
          videotag.load()
          videotag.oncanplay = () => {
            // @ts-ignore
            video.height = videotag.videoHeight
            // @ts-ignore
            video.width = videotag.videoWidth
          }
        })
      }
      reader.readAsDataURL(video.file)
      console.log('done')
    },
    removeFile() {
      const removedFile = this.blog.files[this.blog.files.length - 1]
      const finished = () => {
        this.blog.files.pop()
        this.$toasted.global.success({
          message: `removed file ${removedFile.id}`
        })
      }
      if (this.mode === this.modetypes.add || !removedFile.uploaded) {
        finished()
      } else if (removedFile.name && removedFile.id && this.mode === this.modetypes.edit) {
        this.$axios
          .delete('/deleteFiles', {
            data: {
              fileids: [
                removedFile.id
              ],
              postid: this.blogid,
              posttype: this.type
            }
          })
          .then(res => {
            if (res.status == 200) {
              finished()
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
              message: message
            })
          })
      } else {
        this.$toasted.global.error({
          message: 'no name or id found, or mode type not edit'
        })
      }
    },
    editBlog(searchresult) {
      this.blogid = searchresult.id

      // get images
      const getimages = theblog => {
        let getfilecount = 0
        let gothero = false
        let gottile = false
        let cont = true
        let finished = false
        const finishedGets = () => {
          this.mode = this.modetypes.edit
          this.blog = theblog
          this.$toasted.global.success({
            message: `edit blog with id ${this.blogid}`
          })
        }
        if (theblog.heroimage !== null) {
          axios
            .get(
              `${cloudStorageURLs.blogs}/${
                staticstorageindexes.blogfiles
              }/${this.blogid}/${theblog.heroimage.id + this.paths.original}`,
              {
                responseType: 'blob'
              }
            )
            .then(res => {
              if (!cont) return
              if (res.status == 200) {
                if (res.data) {
                  theblog.heroimage.file = res.data
                  theblog.heroimage.uploaded = true
                  theblog.heroimage.src = null
                  this.updateFileSrc(theblog.heroimage)
                  gothero = true
                  if (
                    theblog.files.length === getfilecount &&
                    gottile && !finished
                  ) {
                    finished = true
                    finishedGets()
                  }
                } else {
                  this.$toasted.global.error({
                    message: 'could not get image data'
                  })
                  cont = false
                }
              } else {
                this.$toasted.global.error({
                  message: `got status code of ${res.status} on image upload`
                })
                cont = false
              }
            })
            .catch(err => {
              this.$toasted.global.error({
                message: `got error on hero image get: ${err}`
              })
              cont = false
            })
        } else {
          theblog.heroimage = Object.assign({}, originalHero)
          theblog.heroimage.id = this.createId()
          gothero = true
          if (
            theblog.files.length === getfilecount &&
            gottile && !finished
          ) {
            finished = true
            finishedGets()
          }
        }
        if (theblog.tileimage !== null) {
          axios
            .get(
              `${cloudStorageURLs.blogs}/${
                staticstorageindexes.blogfiles
              }/${this.blogid}/${theblog.tileimage.id + this.paths.original}`,
              {
                responseType: 'blob'
              }
            )
            .then(res => {
              if (!cont) return
              if (res.status == 200) {
                if (res.data) {
                  theblog.tileimage.uploaded = true
                  theblog.tileimage.file = res.data
                  theblog.tileimage.src = null
                  this.updateFileSrc(theblog.tileimage)
                  gottile = true
                  if (
                    theblog.files.length === getfilecount &&
                    gothero && !finished
                  ) {
                    finished = true
                    finishedGets()
                  }
                } else {
                  this.$toasted.global.error({
                    message: 'could not get image data'
                  })
                  cont = false
                }
              } else {
                this.$toasted.global.error({
                  message: `got status code of ${res.status} on image download`
                })
                cont = false
              }
            })
            .catch(err => {
              this.$toasted.global.error({
                message: `got error on tile image get: ${err}`
              })
              cont = false
            })
        } else {
          theblog.tileimage = Object.assign({}, originalHero)
          theblog.tileimage.id = this.createId()
          gottile = true
          if (
            theblog.files.length === getfilecount &&
            gothero && !finished
          ) {
            finished = true
            finishedGets()
          }
        }
        const getImageFile = filedata => {
          if (!cont) return
          axios
            .get(
              `${cloudStorageURLs.blogs}/${
                staticstorageindexes.blogfiles
              }/${this.blogid}/${filedata.id + this.paths.original}`,
              {
                responseType: 'blob'
              }
            )
            .then(res => {
              if (!cont) return
              if (res.status == 200) {
                if (res.data) {
                  filedata.file = res.data
                  theblog.src = null
                  this.updateFileSrc(filedata)
                  getfilecount++
                  if (
                    theblog.files.length === getfilecount &&
                    gothero && gottile && !finished
                  ) {
                    finished = true
                    finishedGets()
                  }
                } else {
                  this.$toasted.global.error({
                    message: 'could not get image data'
                  })
                  cont = false
                }
              } else {
                this.$toasted.global.error({
                  message: `got status code of ${res.status} on image download`
                })
                cont = false
              }
            })
            .catch(err => {
              this.$toasted.global.error({
                message: `got error on image get: ${err}`
              })
              cont = false
            })
        }
        if (theblog.files.length > 0) {
          for (let i = 0; i < theblog.files.length; i++) {
            if (!cont) break
            theblog.files[i].uploaded = true
            if (this.checkImageType(theblog.files[i].type) && theblog.files[i].type !== 'image/gif')
              getImageFile(theblog.files[i])
            else
              getfilecount++
          }
        } else {
          if (
            gothero && gottile && !finished
          ) {
            finished = true
            finishedGets()
          }
        }
        if (
          theblog.files.length === getfilecount &&
          gothero &&
          gottile &&
          !finished
        ) {
          finished = true
          finishedGets()
        }
      }
      // get blog data first
      this.$apollo.query({query: gql`
        query blog($id: String!, $cache: Boolean!){blog(id: $id, cache: $cache)
        {title, caption, content, id, author, views, heroimage{name, id, width, height, type}, tileimage{name, id, width, height, type}, categories, comments, tags, color, files{name, id, width, height, type}} }
        `, variables: {id: this.id, cache: false}})
        .then(({ data }) => {
          const theblog = data.blog
          getimages(theblog)
        }).catch(err => {
          console.error(err)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    },
    deleteBlog(searchresult) {
      const id = searchresult.id
      this.$apollo.mutate({mutation: gql`
        mutation deleteBlog($id: String!){deleteBlog(id: $id){id} }
        `, variables: {id: id}})
        .then(({ data }) => {
          this.searchresults.splice(
            this.searchresults.indexOf(searchresult),
            1
          )
          this.$toasted.global.success({
            message: 'blog deleted'
          })
        }).catch(err => {
          console.error(err)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    },
    searchblogs(evt) {
      evt.preventDefault()
      this.$apollo.query({query: gql`
        query blogs($perpage: Int!, $page: Int!, $searchterm: String!, $sort: String!, $ascending: Boolean!, $tags: [String!]!, $categories: [String!]!, $cache: Boolean!)
          {blogs(perpage: $perpage, page: $page, searchterm: $searchterm, sort: $sort, ascending: $ascending, tags: $tags, categories: $categories, cache: $cache){title, id} }
        `, variables: {perpage: 10, page: 0, searchterm: this.search, sort: 'title', ascending: false, tags: [], categories: [], cache: false}})
        .then(({ data }) => {
          const blogs = data.blogs
          blogs.map(
            blog => {
              blog.created = this.mongoidToDate(blog.id)
            }
          )
          this.searchresults = blogs
          this.$toasted.global.success({
            message: `found ${this.searchresults.length} result${
              this.searchresults.length === 1 ? '' : 's'
            }`
          })
        }).catch(err => {
          console.error(err)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    },
    clearsearch(evt) {
      if (evt) evt.preventDefault()
      this.search = ''
      this.searchresults = []
    },
    resetblogs(evt) {
      if (evt) evt.preventDefault()
      this.blog = {
        title: '',
        content: '',
        caption: '',
        color: defaultColor,
        author: '',
        heroimage: Object.assign({}, originalHero),
        tileimage: Object.assign({}, originalTile),
        files: [],
        tags: [],
        categories: []
      }
      this.blog.heroimage.id = this.createId()
      this.blog.tileimage.id = this.createId()
      this.mode = this.modetypes.add
      this.blogid = new ObjectID().toString()
    },
    manageblogs(evt) {
      evt.preventDefault()
      let blogid = this.blogid
      this.submitting = true

      // upload image logic
      const upload = () => {
        let cont = true
        let uploadcount = 0
        let fileuploads = this.blog.files.filter(file => !file.uploaded)
        let totaluploads =
          (!this.blog.heroimage.uploaded && this.blog.heroimage.file ? 1 : 0) +
          (!this.blog.tileimage.uploaded && this.blog.tileimage.file ? 1 : 0) +
          fileuploads.length
        let finished = false
        const successMessage = () => {
          this.$toasted.global.success({
            message: `${this.mode}ed blog with id ${blogid}`
          })
          this.submitting = false
          this.resetblogs(evt)
        }
        const uploadFile = (file, fileid) => {
          if (!cont) return
          const formData = new FormData()
          formData.append('file', file)
          this.$axios
            .put('/writeFile', formData, {
              params: {
                posttype: this.type,
                filetype: file.type,
                postid: this.blogid,
                fileid: fileid
              },
              headers: {
                'Content-Type': 'multipart/form-data'
              }
            })
            .then(res => {
              if (!cont) return
              if (res.status == 200) {
                uploadcount++
                if (totaluploads === uploadcount && !finished) {
                  finished = true
                  successMessage()
                }
              } else {
                this.$toasted.global.error({
                  message: `got status code of ${res.status} on file upload`
                })
                cont = false
              }
            })
            .catch(err => {
              let message = `got error: ${err}`
              if (err.response && err.response.data) {
                message = err.response.data.message
              }
              console.log(message)
              this.$toasted.global.error({
                message: message
              })
            })
        }
        let uploadinghero = false
        if (!this.blog.heroimage.uploaded && this.blog.heroimage.file && this.blog.heroimage.type) {
          uploadinghero = true
          this.blog.heroimage.file = new File(
            [this.blog.heroimage.file],
            'hero',
            {
              type: this.blog.heroimage.type
            }
          )
          uploadFile(
            this.blog.heroimage.file,
            this.blog.heroimage.id
          )
        }
        let uploadingtile = false
        if (!this.blog.tileimage.uploaded && this.blog.tileimage.file && this.blog.tileimage.type) {
          uploadingtile = true
          this.blog.tileimage.file = new File(
            [this.blog.tileimage.file],
            'tile',
            {
              type: this.blog.tileimage.type
            }
          )
          uploadFile(
            this.blog.tileimage.file,
            this.blog.tileimage.id
          )
        }
        if (fileuploads.length > 0) {
          for (let i = 0; i < fileuploads.length; i++) {
            fileuploads[i].file = new File(
              [fileuploads[i].file],
              fileuploads[i].name,
              {
                type: fileuploads[i].type
              }
            )
            uploadFile(
              fileuploads[i].file,
              fileuploads[i].id
            )
          }
        }
        if (
          !uploadinghero &&
          fileuploads.length === 0 &&
          !finished
        ) {
          finished = true
          successMessage()
        }
      }

      // send to database logic (do this first)
      const color = this.blog.color.hex8
        ? this.blog.color.hex8
        : this.blog.color.toUpperCase()
      const heroimage = {
        name: 'hero',
        id: this.blog.heroimage.id,
        height: this.blog.heroimage.height,
        width: this.blog.heroimage.width,
        type: this.blog.heroimage.type
      }
      const tileimage = {
        name: 'tile',
        id: this.blog.tileimage.id,
        height: this.blog.tileimage.height,
        width: this.blog.tileimage.width,
        type: this.blog.tileimage.type
      }
      const files = this.blog.files.map(file => () => {
        return {
          id: file.id,
          name: file.name,
          width: file.width ? file.width : 0,
          height: file.height ? file.height : 0,
          type: file.type
        }
      })
      if (this.mode === this.modetypes.add) {
        this.$apollo.mutate({mutation: gql`
          mutation addBlog($id: String!, $title: String!, $content: String!, $color: String!, $caption: String!, $author: String!, $tileimage: FileInput!, $heroimage: FileInput!, $files: [FileInput!]! $tags: [String!]!, $categories: [String!]!)
          {addBlog(id: $id, title: $title, content: $content, color: $color, caption: $caption, author: $author, tileimage: $tileimage, heroimage: $heroimage, files: $files, tags: $tags, categories: $categories,){id} }
          `, variables: {id: this.blogid, title: this.blog.title, content: this.blog.content, color, caption: this.blog.caption, author: this.blog.author, tileimage, heroimage, files, categories: this.blog.categories, tags: this.blog.tags}})
          .then(({ data }) => {
            blogid = data.addBlog.id
            upload()
          }).catch(err => {
            console.error(err)
            this.$toasted.global.error({
              message: `found error: ${err.message}`
            })
          })
      } else {
        this.$apollo.mutate({mutation: gql`
          mutation updateBlog($id: String!, $title: String!, $content: String!, $color: String!, $caption: String!, $author: String!, $tileimage: FileInput!, $heroimage: FileInput!, $files: [FileInput!]! $tags: [String!]!, $categories: [String!]!)
          {updateBlog(id: $id, title: $title, content: $content, color: $color, caption: $caption, author: $author, tileimage: $tileimage, heroimage: $heroimage, files: $files, tags: $tags, categories: $categories,){id} }
          `, variables: {id: this.blogid, title: this.blog.title, content: this.blog.content, color, caption: this.blog.caption, author: this.blog.author, tileimage, heroimage, files, categories: this.blog.categories, tags: this.blog.tags}})
          .then(({ data }) => {
            upload()
          }).catch(err => {
            console.error(err)
            this.$toasted.global.error({
              message: `found error: ${err.message}`
            })
          })
      }
    }
  }
})
</script>

<style lang="scss">
@import '~/node_modules/prismjs/themes/prism.css';
.arrow-size-edit {
  font-size: 1rem;
}
.markdown {
  overflow: auto;
  max-height: 20rem;
}
.sampleimage {
  max-width: 200px;
}
</style>
