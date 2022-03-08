export function groupBy<K, V>(arr: Array<V>, fn: (v: V) => K): Array<[K, Array<V>]> {
  const groups = arr.reduce((map, value) => {
    const key = fn(value);
    const group = map.get(key) || [];
    group.push(value);
    map.set(key, group);
    return map;
  }, new Map as Map<K, Array<V>>);

  return Array.from(groups.entries());
}
