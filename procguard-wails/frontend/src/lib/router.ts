import { writable } from "svelte/store";

export const currentPath = writable(window.location.pathname);

export function navigate(path: string) {
  window.history.pushState({}, "", path);
  currentPath.set(path);
}
