import * as preact from "preact"
import * as vlens from "vlens";
import * as events from "vlens/events"
import * as server from "@app/server";

export async function fetch(route: string, prefix: string) {
    return server.ListUsers({})
}

type Form = {
    data: server.UserListResponse
    name: string
    error: string
}

const useForm = vlens.declareHook((data: server.UserListResponse): Form => ({
    data, name: "", error: ""
}))

export function view(route: string, prefix: string, data: server.UserListResponse): preact.ComponentChild {
    let form = useForm(data)
    return <div>
        <h3>Users</h3>
        {form.data.AllUsernames.map(name => <div key={name}>{name}</div>)}
        <h3>Add User</h3>
        <input type="text" {...events.inputAttrs(vlens.ref(form, "name"))} />
        <button onClick={vlens.cachePartial(onAddUserClicked, form)}>Add</button>
        {form.name && <div>
            You are inputting: <code>{form.name}</code>
        </div>}
    </div>
}

async function onAddUserClicked(form: Form) {
    let [resp, err] = await server.AddUser({Username: form.name})
    if (resp) {
        form.name = ""
        form.data = resp
        form.error = ""
    } else {
        form.error = err
    }
    vlens.scheduleRedraw()
}
