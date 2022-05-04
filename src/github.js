import fetch from 'node-fetch';

export default async () => {
  const user = process.argv[2]
  const url = `https://api.github.com/users/${user}/events`;
  const res = await fetch(url).then(response =>  response.json());
  const today = new Date();

  let count = 0;

  res.forEach(e => {
    const created_at = new Date(e.created_at);
    if (created_at.getFullYear() !== today.getFullYear()) return;
    if (created_at.getMonth() !== today.getMonth()) return;
    if (created_at.getDate() !== today.getDate()) return;
    if (e.type == 'PushEvent') { count += 1 }
  });
  return count;
};
