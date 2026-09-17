import DefaultTheme from 'vitepress/theme'
import { h, onMounted, watch, nextTick } from 'vue'
import { useRoute } from 'vitepress'
import mediumZoom from 'medium-zoom'
import './style.css'
import ContactQR from './components/ContactQR.vue'
import DocStepFlow from './components/DocStepFlow.vue'

export default {
    extends: DefaultTheme,
    Layout: () => {
        return h(DefaultTheme.Layout, null, {
            'nav-bar-content-after': () => h(ContactQR)
        })
    },
    enhanceApp({ app }: { app: any }) {
        app.component('DocStepFlow', DocStepFlow)
    },
    setup() {
        const route = useRoute()
        const initZoom = () => {
            // 对正文区所有图片启用全屏无级缩放灯箱（排除头像等特定图标）
            mediumZoom('.vp-doc img:not(.no-zoom)', {
                background: 'rgba(11, 15, 23, 0.92)',
                margin: 24
            })
        }

        onMounted(() => {
            initZoom()
        })

        watch(
            () => route.path,
            () => nextTick(() => initZoom())
        )
    }
}
