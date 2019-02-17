//Formタグを作らずにJavascriptのみで通常のPOST送信
//https://gist.github.com/zuzu/c68e105d966c4d235334
function execPost(a, data) {

    if (a == "update") {
        data.name = document.getElementById("id_" + data.id).value;
        action = "/category_update";
    } else if (a == "delete") {
        let result = window.confirm("削除しますか？");
        if (!result) {
            return;
        }
        action = "/category_delete";
    } else {
        return;
    }

    // フォームの生成
    let form = document.createElement("form");
    form.setAttribute("action", action);
    form.setAttribute("method", "post");
    form.style.display = "none";
    document.body.appendChild(form);
    // パラメタの設定
    if (data !== undefined) {
        for (var paramName in data) {
            var input = document.createElement('input');
            input.setAttribute('type', 'hidden');
            input.setAttribute('name', paramName);
            input.setAttribute('value', data[paramName]);
            form.appendChild(input);
        }
    }
    // submit
    form.submit();
}
