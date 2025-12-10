import { w as writable } from "./index.js";
function createMapStore() {
  const { subscribe, set, update } = writable({
    data: null,
    lastUpdate: 0
  });
  return {
    subscribe,
    setMapData: (data) => {
      update((store) => ({
        ...store,
        data,
        lastUpdate: Date.now()
      }));
    },
    clear: () => set({
      data: null,
      lastUpdate: 0
    })
  };
}
createMapStore();
