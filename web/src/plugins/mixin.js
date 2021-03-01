import Vue from "vue";

function gDateFormat(p) {
    var da = new Date(p);
    var year = da.getFullYear();
    var month = da.getMonth() + 1;
    var dt = da.getDate();
    var h = da.getHours();
    var m = da.getMinutes();
    var s = da.getSeconds();

    return year + '-' + month + '-' + dt + ' ' + h + ':' + m + ':' + s;
}

var Mixin = {
    data() {
        return {
            user_edit_dialog: false,
            isLoading: false,
        }
    },
    computed: {},
    methods: {
        tableDateFormat(row, column) {
            var p = row[column.property];
            if (p === undefined) {
                return "";
            }
            return gDateFormat(p);
        },
        tableArrayFormat(row, column) {
            var p = row[column.property];
            if (p === undefined) {
                return "";
            }
            return p.join("\n\r\n");
        },
        disVisible() {
            this.user_edit_dialog = false
        },
    },
}

// Vue.filter("dateFormat", function (p) {
//     return gDateFormat(p);
// })
Vue.mixin(Mixin)


// export default Mixin

