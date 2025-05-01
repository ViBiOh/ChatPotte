let authToken = "";
let idUser = "";

async function clearMessages() {
  const channel = window.location.href.split("/").pop();
  const baseURL = `https://discordapp.com/api/channels/${channel}/messages`;
  const headers = { Authorization: authToken };

  let before = new Date();
  before.setMonth(before.getMonth() - 2);

  function wait(duration) {
    return new Promise((resolve) => {
      setTimeout(resolve, duration);
    });
  }

  function getMessages(before) {
    let url = `${baseURL}?limit=100`;
    if (before) {
      url += `&before=${before}`;
    }

    return fetch(url, { headers });
  }

  function deleteMessage(message) {
    console.log(message);

    return fetch(`${baseURL}/${message.id}`, { headers, method: "DELETE" });
  }

  function filterMessages(message) {
    return (
      (idUser === message.author.id ||
        (message.author.bot && message.content.indexOf(idUser) != -1)) &&
      new Date(message.timestamp) < before
    );
  }

  let beforeID;

  while (true) {
    messages = await getMessages(beforeID);
    messages = await messages.json();

    if (messages.length === 0) {
      console.info(`Done for ${channel}`);
      return;
    }

    console.info(`Fetched ${messages.length} messages`);

    beforeID = messages[messages.length - 1].id;
    messages = messages.filter(filterMessages);

    if (messages.length > 0) {
      console.info(`Cleaning ${messages.length} messages...`);
    }

    for (var i = 0; i < messages.length; i++) {
      await wait(1500);

      let resp = await deleteMessage(messages[i]);
      if (resp && resp.status === 204) {
        continue;
      }

      console.error(resp);
    }
  }
}
