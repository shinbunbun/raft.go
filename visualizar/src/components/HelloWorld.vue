<template>
  <div class="hello container">
    <div class="row">
      <div class="col text-center">
        <table class="table fs-2">
          <thead>
            <tr>
              <th scope="col">Leader</th>
              <th scope="col">Term</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <th scope="row">{{ Leader }}</th>
              <td>{{ Term }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div class="row">
      <div v-for="(d, i) in data" :key="i" class="col">
        <div class="card col-3" style="width: 18rem; margin-bottom: 20px">
          <div class="card-body">
            <h5 class="card-title">{{ d.Node.name }}</h5>
            <p class="card-text">{{ d.Log }}</p>
          </div>
        </div>
      </div>
    </div>
    <div class="row" style="margin: 30px">
      <div class="col">
        <label for="textttttt">Append Log: </label>
        <input type="text" name="textttttt" id="text" />
        <button
          class="btn btn-primary"
          style="margin-left: 10px"
          @click="appendLog"
        >
          Send
        </button>
      </div>
      <div class="col"></div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'HelloWorld',
  props: {
    msg: String
  },
  async mounted() {
    console.log('Component mounted.')
    let response, bodyJson;
    try {
     response = await fetch('http://localhost:8080/monitor?address=172.26.250.11:8000')
     bodyJson = await response.json()
      console.log(bodyJson)
      this.members = bodyJson.members
      this.members.push(bodyJson.node_info)
    } catch (e) {
      console.log(e)
    }
    setInterval(() => {
      this.getInfo()
    }, 1000);
  },
  data() {
    return {
      members: [],
      data: {},
      Leader: '',
      Term: -1,
      input: ''
    }
  },
  methods: {
    async getInfo() {
        this.members.forEach( async e => {
          try {
            const response = await fetch('http://localhost:8080/info?address=' + e.endpoint)
            const bodyJson = await response.json()
            console.log(bodyJson)
            this.data[bodyJson.Node.name] = bodyJson
            this.Leader = bodyJson.Leader
            this.Term = bodyJson.Term
          } catch (e) {
            console.log(e)
          }
        });
    },
    async appendLog() {
      try {
        const log = document.getElementById('text').value
        await fetch('http://localhost:8080/append?address=' + this.data[this.Leader].Node.endpoint + '&log=' + log, {
          method: 'POST'
        })
        document.getElementById('text').value = ''
      } catch (e) {
        console.log(e)
      }
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h3 {
  margin: 40px 0 0;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
