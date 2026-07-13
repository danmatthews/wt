// Keeps the shadcn-vue `.dark` class on <html> in sync with the OS appearance,
// updating live when the user flips their system light/dark setting.
export function followSystemTheme() {
  const query = window.matchMedia("(prefers-color-scheme: dark)");

  const apply = () =>
    document.documentElement.classList.toggle("dark", query.matches);

  apply();
  query.addEventListener("change", apply);
}
