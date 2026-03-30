import DefaultTheme from 'vitepress/theme'
import { h } from 'vue'
import './style.css'
import ContactQR from './components/ContactQR.vue'

export default {
    extends: DefaultTheme,
    Layout: () => {
        return h(DefaultTheme.Layout, null, {
            'nav-bar-content-after': () => h(ContactQR)
        })
    }
}
