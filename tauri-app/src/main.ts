import { createApp } from "vue";
import "./assets/index.css";
import App from "./App.vue";
import { followSystemTheme } from "./lib/theme";

followSystemTheme();
createApp(App).mount("#app");
