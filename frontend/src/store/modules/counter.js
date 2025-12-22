import { defineStore } from 'pinia';

export const useCounterStore = defineStore('counter', {
  state: () => ({
    count: 0,
  }),
  getters: {
    doubleCount: (state) => state.count * 2,
    isEven: (state) => state.count % 2 === 0,
  },
  actions: {
    accumulate() {
      this.count++;
    },
    decrement() {
      this.count--;
    },
    reset() {
      this.count = 0;
    },
    incrementBy(value) {
      this.count += value;
    },
    decrementBy(value) {
      this.count -= value;
    },
  },
});
