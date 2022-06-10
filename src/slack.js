import { WebClient } from '@slack/web-api';

export default (count) => {
  let message = '';
  if (count == 0) {
    message = `<!channel> 今日はまだコミットをしていません!\ñhttps://github.com/${process.argv[2]}`;
  } else {
    message = `今日のコミット数は${count}\ñhttps://github.com/${process.argv[2]}`;
  }

  const token = process.argv[3];
  const web = new WebClient(token);
  const conversationId = process.argv[4];

  (async () => {
    const res = await web.chat.postMessage({ channel: conversationId, text: message });
    console.log('Message sent: ', res.ts);
  })();
};
