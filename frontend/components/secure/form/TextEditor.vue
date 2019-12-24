<template>
  <div>
    <span @mouseover="focusMenu = true" @mouseleave="focusMenu = false">
      <editor-menu-bar
        v-if="showMenu && (focusMenu || focusEditor)"
        :editor="editor"
        v-slot="{ commands, isActive }"
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
          >
            H1
          </b-nav-item>
          <b-nav-item
            :active="isActive.heading({ level: 2 })"
            @click="commands.heading({ level: 2 })"
          >
            H2
          </b-nav-item>
          <b-nav-item
            :active="isActive.heading({ level: 3 })"
            @click="commands.heading({ level: 3 })"
          >
            H3
          </b-nav-item>
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
          <b-nav-item @click="showImagePrompt(commands.image)">
            <client-only>
              <font-awesome-icon icon="image" />
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
    </span>
    <editor-content
      :editor="editor"
      :style="{
        'margin-top': showMenu
          ? focusMenu || focusEditor
            ? '1rem'
            : '6rem'
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
  CodeBlockHighlight,
  Image
} from 'tiptap-extensions'
import javascript from 'highlight.js/lib/languages/javascript'
import css from 'highlight.js/lib/languages/css'
import go from 'highlight.js/lib/languages/go'
import java from 'highlight.js/lib/languages/java'
import cpp from 'highlight.js/lib/languages/cpp'
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
      focusMenu: false,
      focusEditor: false,
      editor: new Editor({
        onFocus: () => {
          this.focusEditor = true
        },
        onBlur: () => {
          this.focusEditor = false
        },
        onUpdate: () => {
          const data = this.editor.getJSON()
          this.$emit('updated-text', data)
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
          new Image(),
          new History()
        ],
        content: 'text goes here'
      })
    }
  },
  beforeDestroy() {
    this.editor.destroy()
  }
})
</script>

<style lang="scss"></style>
