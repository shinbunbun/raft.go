<template>
  <MyHeader/>
  <HelloWorld msg="Welcome to Your Vue.js App"/>
</template>

<script>
import HelloWorld from './components/HelloWorld.vue'
import MyHeader from './components/MyHeader.vue';

export default {
  name: 'App',
  components: {
    HelloWorld,
    MyHeader
}
}
</script>
