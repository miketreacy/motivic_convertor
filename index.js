const apiConfig = {
    url: `/upload/midi`,
    method: 'POST',
    mode: 'cors'
};

const baseURL = window.location;
const formEl = document.querySelector("#file-upload");
const spinnerEl = formEl.querySelector('#spinner');
const uploadBtn = formEl.querySelector('#upload');
const outputNameEl = formEl.querySelector("#output-name");
const waveFormEl = formEl.querySelector("#waveform");
const fileInputEl = formEl.querySelector("#upload-file");
const loadingIcon = `&#8635;`;
const messages = {
    "arrow-up": `&#8679;`,
    "arrow-down": `&#8681;`
};

function getApiParams(payload) {
    let { method, mode } = apiConfig;
    return {
        method,
        body: payload,
        mode
    }

}

function getFetchArgs(payload) {
    return [apiConfig.url, getApiParams(payload)];
}

async function awaitFetch(url, params) {
    try {
        let res = await window.fetch(url, params);
        return await res.json();
    } catch (e) {
        console.error(e);
    }
}

function fileInputChange(e) {
    uploadBtn.disabled = !fileInputEl.files.length;
}

function loading(btnEl, loading = true) {
    // TODO: make loadingIcon spin!
    const iconEls = btnEl.querySelectorAll('.icon');
    if (loading) {
        iconEls.forEach(el => el.innerHTML = loadingIcon);
    } else {
        iconEls.forEach(el => el.innerHTML = messages[el.dataset.icon]);
    }
}


function displayDownloadButon(url) {
    const downloadBtn = document.querySelector('#download');
    downloadBtn.href = url;
    downloadBtn.addEventListener('click', e => {
        downloadBtn.classList.add('hide');
        uploadBtn.classList.remove('hide');
    });
    uploadBtn.classList.add('hide');
    downloadBtn.classList.remove('hide');
}


async function uploadClick(e) {
    loading(uploadBtn);
    const files = fileInputEl.files;
    const formData = new FormData();
    formData.append('myMIDIFile', files[0]);
    formData.append(outputNameEl.name, outputNameEl.value);
    formData.append(waveFormEl.name, waveFormEl.value);
    let data = await awaitFetch(...getFetchArgs(formData));
    console.log('API response from /upload/midi...');
    console.dir(data);
    loading(uploadBtn, false);
    if (data && data.url) {
        displayDownloadButon(data.url);
    } else {
        window.alert(data.message)
    }
}


fileInputEl.addEventListener('change', fileInputChange)
uploadBtn.addEventListener('click', uploadClick);