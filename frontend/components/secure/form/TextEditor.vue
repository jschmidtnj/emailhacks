<template>
  <div class="editor">
    <editor-menu-bar
      v-if="showMenu"
      :editor="editor"
      v-slot="{ commands, isActive }"
    >
      <b-nav pills class="menubar">
        <b-nav-item :active="isActive.bold()" @click="commands.bold">
          <client-only>
            <font-awesome-icon class="mr-2" icon="bold" />
          </client-only>
        </b-nav-item>
        <b-nav-item :active="isActive.italic()" @click="commands.italic">
          <client-only>
            <font-awesome-icon class="mr-2" icon="italic" />
          </client-only>
        </b-nav-item>
        <b-nav-item :active="isActive.strike()" @click="commands.strike">
          <client-only>
            <font-awesome-icon class="mr-2" icon="strikethrough" />
          </client-only>
        </b-nav-item>
        <b-nav-item :active="isActive.underline()" @click="commands.underline">
          <client-only>
            <font-awesome-icon class="mr-2" icon="underline" />
          </client-only>
        </b-nav-item>
        <b-nav-item :active="isActive.code()" @click="commands.code">
          <client-only>
            <font-awesome-icon class="mr-2" icon="code" />
          </client-only>
        </b-nav-item>
        <b-nav-item :active="isActive.paragraph()" @click="commands.paragraph">
          <client-only>
            <font-awesome-icon class="mr-2" icon="paragraph" />
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
            <font-awesome-icon class="mr-2" icon="list-ul" />
          </client-only>
        </b-nav-item>
        <b-nav-item
          :active="isActive.ordered_list()"
          @click="commands.ordered_list"
        >
          <client-only>
            <font-awesome-icon class="mr-2" icon="list-ol" />
          </client-only>
        </b-nav-item>
        <b-nav-item
          :active="isActive.blockquote()"
          @click="commands.blockquote"
        >
          <client-only>
            <font-awesome-icon class="mr-2" icon="quote-right" />
          </client-only>
        </b-nav-item>
        <b-nav-item
          :active="isActive.code_block()"
          @click="commands.code_block"
        >
          <client-only>
            <font-awesome-icon class="mr-2" icon="code" />
          </client-only>
        </b-nav-item>
        <b-nav-item @click="commands.horizontal_rule">
          <client-only>
            <font-awesome-icon class="mr-2" icon="grip-lines" />
          </client-only>
        </b-nav-item>
        <b-nav-item @click="commands.undo">
          <client-only>
            <font-awesome-icon class="mr-2" icon="undo" />
          </client-only>
        </b-nav-item>
        <b-nav-item @click="commands.redo">
          <client-only>
            <font-awesome-icon class="mr-2" icon="redo" />
          </client-only>
        </b-nav-item>
      </b-nav>
    </editor-menu-bar>
    <editor-content :editor="editor" class="mt-4" />
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import { Editor, EditorContent, EditorMenuBar } from 'tiptap'
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
} from 'tiptap-extensions'

export default Vue.extend({
  components: {
    EditorContent,
    EditorMenuBar
  },
  props: {
    showMenu: {
      default: true,
      type: Boolean
    }
  },
  data() {
    return {
      editor: new Editor({
        extensions: [
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
        ],
        content: 'text goes here',
      }),
    }
  },
  beforeDestroy() {
    this.editor.destroy()
  },
})
</script>
