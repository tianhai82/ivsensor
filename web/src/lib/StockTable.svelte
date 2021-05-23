<script>
  export let stockAtr = { OptionPremiums: [] };
</script>

<div class="w-full overflow-x-auto">
  <table
    class="table-auto w-full border-collapse border border-teal-300 text-sm lg:text-base"
  >
    <thead>
      <tr>
        <th>Ticker</th>
        <th>Stock Price</th>
        <th>Weekly ATR</th>
        <th>Normalized ATR</th>
        <th>DTE</th>
        <th>Expiry Date</th>
        <th>Put Strike</th>
        <th>Put Premium</th>
        <th>Put Premium Annualized</th>
      </tr>
    </thead>
    <tbody>
      {#each stockAtr.OptionPremiums as premium}
        <tr>
          <td>{stockAtr.Symbol}</td>
          <td>{stockAtr.CurrentStockPrice.toFixed(2)}</td>
          <td>{stockAtr.WeeklyATR.toFixed(2)}</td>
          <td>{premium.NormalizedATR.toFixed(2)}</td>
          <td>{premium.DTE}</td>
          <td>{premium.ExpiryDate.substr(0, 10)}</td>
          <td>{premium.PutStrike}</td>
          <td>{premium.PutPremium.toFixed(3)}</td>
          <td>{(premium.PutPremiumAnnualizedPercent * 100).toFixed(1)}%</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<style>
  th {
    @apply border border-green-600;
  }
  td {
    @apply border border-green-600 text-center;
  }
</style>
