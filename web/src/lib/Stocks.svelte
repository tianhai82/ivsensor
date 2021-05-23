<script>
  import { Autocomplete, Spinner, Input } from "svetamat2/src";
  import StockTable from "./StockTable.svelte";

  let tickers = [];
  let stock1;
  let stock2;

  fetch("/api/web/tickers")
    .then((resp) => {
      if (resp.status !== 200) {
        throw {
          status: resp.status,
          statusText: resp.statusText,
        };
      }
      return resp.json();
    })
    .then((t) => (tickers = t))
    .catch((err) => console.log(err));

  let stockAtr1 = { OptionPremiums: [] };
  let stockAtr2 = { OptionPremiums: [] };
  let matchScore = "";

  let p1, p2, scorePromise;
  function analyseStock1({ detail }) {
    p1 = analyseStock(detail).then((stockAtr) => {
      stockAtr1 = stockAtr;
    });
  }

  function analyseStock2({ detail }) {
    p2 = analyseStock(detail).then((stockAtr) => {
      stockAtr2 = stockAtr;
    });
  }

  $: if (stock1 && stock2) {
    scorePromise = pairScore(stock1, stock2).then((score) => {
      matchScore = score.toFixed(2);
    });
  }

  function pairScore(ticker1, ticker2) {
    return fetch(`/api/web/pairscore/${ticker1}/${ticker2}`)
      .then((resp) => {
        if (resp.status !== 200) {
          throw {
            status: resp.status,
            statusText: resp.statusText,
          };
        }
        return resp.json();
      })
      .catch((err) => console.log(err));
  }

  function analyseStock(ticker) {
    return fetch(`/api/web/analyse/${ticker}`)
      .then((resp) => {
        if (resp.status !== 200) {
          throw {
            status: resp.status,
            statusText: resp.statusText,
          };
        }
        return resp.json();
      })
      .then((atr) => {
        if (!atr.OptionPremiums) {
          atr.OptionPremiums = [];
        }
        return atr;
      })
      .catch((err) => console.log(err));
  }
</script>

<div class="h-full w-full flex flex-col gap-y-2 p-2 overflow-auto">
  <div
    class="flex-1 elevation-4 rounded-lg bg-white flex flex-col gap-y-3 px-2 pb-2"
  >
    <Autocomplete
      hideDetails
      label="Stock Ticker"
      items={tickers}
      bind:value={stock1}
      minCharactersToSearch={1}
      on:change={analyseStock1}
    />
    {#await p1}
      <div class="flex gap-x-2">
        <Spinner /> Analysing {stock1}
      </div>
    {:then}
      {#if stock1}
        <StockTable stockAtr={stockAtr1} />
      {/if}
    {/await}
  </div>
  <div
    class="flex-1 elevation-4 rounded-lg bg-white flex flex-col gap-y-3 px-2 pb-2"
  >
    <Autocomplete
      hideDetails
      label="Stock Ticker"
      items={tickers}
      bind:value={stock2}
      minCharactersToSearch={1}
      on:change={analyseStock2}
    />
    {#await p2}
      <div class="flex gap-x-2">
        <Spinner /> Analysing {stock2}
      </div>
    {:then}
      {#if stock2}
        <StockTable stockAtr={stockAtr2} />
      {/if}
    {/await}

    {#await scorePromise}
      <div class="flex gap-x-2">
        <Spinner /> Calculating pair score for {stock1} and {stock2}
      </div>
    {:then}
      {#if matchScore}
        <div class="">
          <Input
            hideDetails
            outlined
            label="{stock1} vs {stock2} (> than 0.5 = negative correlated)"
            value={matchScore}
            readonly
          />
        </div>
      {/if}
    {/await}
  </div>
</div>
