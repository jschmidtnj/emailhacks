<template>
  <div>
    <span @mouseover="focusMenu = true" @mouseleave="focusMenu = false">
      <editor-menu-bar
        v-if="showMenu && (focusMenu || focusEditor)"
        v-slot="{ commands, isActive }"
        :editor="editor"
      >
        <b-nav pills class="menubar">
          <b-nav-item :active="isActive.bold()" @click="commands.bold">
            <client-only>
              <font-awesome-icon icon="bold" />
            </client-only>
          </b-nav-item>
          <b-nav-item :active="isActive.italic()" @click="commands.italic">
            <client-only>
              <font-awesome-icon icon="italic" />
            </client-only>
          </b-nav-item>
          <b-nav-item :active="isActive.strike()" @click="commands.strike">
            <client-only>
              <font-awesome-icon icon="strikethrough" />
            </client-only>
          </b-nav-item>
          <b-nav-item
            :active="isActive.underline()"
            @click="commands.underline"
          >
            <client-only>
              <font-awesome-icon icon="underline" />
            </client-only>
          </b-nav-item>
          <b-nav-item
            :active="isActive.emoji()"
            @click="
              () => {
                const key =
                  emojiData.ordered[
                    Math.floor(Math.random() * emojiData.ordered.length)
                  ]
                commands.emoji({
                  id: key,
                  label: emojiData.lib[key].char
                })
              }
            "
          >
            <client-only>
              <font-awesome-icon icon="smile" />
            </client-only>
          </b-nav-item>
          <b-nav-item :active="isActive.code()" @click="commands.code">
            <client-only>
              <font-awesome-icon icon="code" />
            </client-only>
          </b-nav-item>
          <b-nav-item
            :active="isActive.paragraph()"
            @click="commands.paragraph"
          >
            <client-only>
              <font-awesome-icon icon="paragraph" />
            </client-only>
          </b-nav-item>
          <b-nav-item
            :active="isActive.heading({ level: 1 })"
            @click="commands.heading({ level: 1 })"
            >H1</b-nav-item
          >
          <b-nav-item
            :active="isActive.heading({ level: 2 })"
            @click="commands.heading({ level: 2 })"
            >H2</b-nav-item
          >
          <b-nav-item
            :active="isActive.heading({ level: 3 })"
            @click="commands.heading({ level: 3 })"
            >H3</b-nav-item
          >
          <b-nav-item
            :active="isActive.bullet_list()"
            @click="commands.bullet_list"
          >
            <client-only>
              <font-awesome-icon icon="list-ul" />
            </client-only>
          </b-nav-item>
          <b-nav-item
            :active="isActive.ordered_list()"
            @click="commands.ordered_list"
          >
            <client-only>
              <font-awesome-icon icon="list-ol" />
            </client-only>
          </b-nav-item>
          <b-nav-item
            :active="isActive.blockquote()"
            @click="commands.blockquote"
          >
            <client-only>
              <font-awesome-icon icon="quote-right" />
            </client-only>
          </b-nav-item>
          <b-nav-item
            :active="isActive.code_block()"
            @click="commands.code_block"
          >
            <client-only>
              <font-awesome-icon icon="code" />
            </client-only>
          </b-nav-item>
          <b-nav-item @click="commands.horizontal_rule">
            <client-only>
              <font-awesome-icon icon="grip-lines" />
            </client-only>
          </b-nav-item>
          <b-nav-item @click="commands.undo">
            <client-only>
              <font-awesome-icon icon="undo" />
            </client-only>
          </b-nav-item>
          <b-nav-item @click="commands.redo">
            <client-only>
              <font-awesome-icon icon="redo" />
            </client-only>
          </b-nav-item>
        </b-nav>
      </editor-menu-bar>
      <div v-else-if="showMenu">
        <hr style="margin-top:3rem;" />
      </div>
    </span>
    <editor-content
      :editor="editor"
      :style="{
        'margin-top': showMenu
          ? focusMenu || focusEditor
            ? '1rem'
            : '3.5rem'
          : '15px'
      }"
      class="editor__content"
    />
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import { Editor, EditorMenuBar, EditorContent } from 'tiptap'
import {
  Blockquote,
  CodeBlock,
  HardBreak,
  Heading,
  HorizontalRule,
  OrderedList,
  BulletList,
  ListItem,
  TodoItem,
  TodoList,
  Bold,
  Code,
  Italic,
  Link,
  Strike,
  Underline,
  History,
  CodeBlockHighlight
} from 'tiptap-extensions'
import emojiData from 'emojilib'
import javascript from 'highlight.js/lib/languages/javascript'
import css from 'highlight.js/lib/languages/css'
import go from 'highlight.js/lib/languages/go'
import java from 'highlight.js/lib/languages/java'
import cpp from 'highlight.js/lib/languages/cpp'
import Emoji from '~/assets/Emoji'
// tried to get emojis working, following this:
// https://github.com/scrumpy/tiptap/blob/master/examples/Components/Routes/Suggestions/index.vue
// turns out the tooltip doesn't want to work:
// https://github.com/Human-Connection/Human-Connection/pull/2258
// currently looking for alternatives
// look into mathjax for math editing: https://www.npmjs.com/package/vue-mathjax
export default Vue.extend({
  components: {
    EditorMenuBar,
    EditorContent
  },
  props: {
    showMenu: {
      type: Boolean,
      default: true
    }
  },
  data() {
    return {
      emojiData,
      focusMenu: false,
      focusEditor: false,
      query: null,
      emojiSuggestionRange: null,
      filteredEmojis: [],
      insertEmoji: () => {},
      editor: new Editor({
        onFocus: () => {
          this.focusEditor = true
        },
        onBlur: () => {
          this.focusEditor = false
        },
        onUpdate: () => {
          const data = this.editor.getHTML()
          this.$emit('updated-text', data)
        },
        onDrop(view, event, slice, moved) {
          // return true to stop the drop event
          // this will just prevent drop from external sources
          return !moved;
        },
        extensions: [
          new CodeBlockHighlight({
            languages: {
              javascript,
              css,
              go,
              java,
              cpp
            }
          }),
          new Blockquote(),
          new BulletList(),
          new CodeBlock(),
          new HardBreak(),
          new Heading({ levels: [1, 2, 3] }),
          new HorizontalRule(),
          new ListItem(),
          new OrderedList(),
          new TodoItem(),
          new TodoList(),
          new Link(),
          new Bold(),
          new Code(),
          new Italic(),
          new Strike(),
          new Underline(),
          new History(),
          new Emoji({
            // a list of all suggested items
            items: () => {
              return this.emojiData.ordered.map(key => {
                return {
                  id: key,
                  data: this.emojiData.lib[key].char
                }
              })
            },
            // is called when a suggestion starts
            onEnter: ({
              items, query, range, command, decorationNode, virtualNode,
            }) => {
              this.query = query
              this.filteredEmojis = items
              this.emojiSuggestionRange = range
              // we save the command for inserting a selected emoji
              // via keyboard navigation and on click
              this.insertEmoji = command
            },
            // is called when a suggestion has changed
            onChange: ({
              items, query, range, decorationNode, virtualNode
            }) => {
              this.query = query
              this.filteredEmojis = items
              this.emojiSuggestionRange = range
            },
            // is called when a suggestion is cancelled
            onExit: () => {
              // reset all saved values
              this.query = null
              this.filteredEmojis = []
              this.emojiSuggestionRange = null
            },
            // is called on every keyDown event while a suggestion is active
            onKeyDown: ({ event }) => {
              // pressing enter
              if (event.keyCode === 13) {
                if (this.hasResults) {
                  this.selectEmoji(this.filteredEmojis[0])
                }
                return true
              }
              return false
            }
          })
        ],
        content: 'text goes here'
      })
    }
  },
  computed: {
    hasResults() {
      return this.filteredEmojis.length
    },
    showSuggestions() {
      return this.query || this.hasResults
    },
  },
  beforeDestroy() {
    this.editor.destroy()
  },
  methods: {
    // we have to replace our suggestion text with a emoji
    // so it's important to pass also the position of your suggestion text
    selectEmoji(emoji) {
      this.insertEmoji({
        range: this.emojiSuggestionRange,
        attrs: {
          id: emoji.id,
          label: emoji.data,
        },
      })
      this.editor.focus()
    }
  }
})
</script>

<style lang="scss">
.emoji {
  font-size: 0.8rem;
  font-weight: bold;
  border-radius: 5px;
  padding: 0.2rem 0.5rem;
  white-space: nowrap;
}
</style>
