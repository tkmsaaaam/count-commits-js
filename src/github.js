import fetch from 'node-fetch';

export default async () => {
  const user = process.argv[2]
  const url = `https://api.github.com/users/${user}/events`;
  const res = await fetch(url).then(response =>  response.json());
  const today = new Date().toLocaleString({ timeZone: 'Asia/Tokyo' });

  let count = 0;

  res.forEach(e => {
    const created_at = new Date(e.created_at).toLocaleString({ timeZone: 'Asia/Tokyo' });
    if (today.slice(0,9) == created_at.slice(0,9) && e.type == 'PushEvent') { count += 1 }
  });
  return count;
};
